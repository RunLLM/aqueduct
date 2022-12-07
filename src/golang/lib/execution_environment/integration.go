package execution_environment

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	collection_utils "github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	CondaPathKey = "conda_path"
	ExecStateKey = "exec_state"
)

func serializeExecStateAndLogFailure(execState *shared.ExecutionState) string {
	serializedState, err := json.Marshal(execState)
	if err != nil {
		// We should never hit this
		log.Errorf("Error marshalling serialized state: %v", err)
		return ""
	}

	return string(serializedState)
}

func serializedFailure(
	outputs string,
	msg string,
	runningAt *time.Time,
) string {
	failureType := shared.SystemFailure
	now := time.Now()
	execState := &shared.ExecutionState{
		Status:      shared.FailedExecutionStatus,
		FailureType: &failureType,
		UserLogs: &shared.Logs{
			StdErr: outputs,
		},
		Error: &shared.Error{
			Context: msg,
			Tip:     shared.TipUnknownInternalError,
		},
		Timestamps: &shared.ExecutionTimestamps{
			RunningAt:  runningAt,
			FinishedAt: &now,
		},
	}

	return serializeExecStateAndLogFailure(execState)
}

func serializedRunning(runningAt *time.Time) string {
	execState := &shared.ExecutionState{
		Status: shared.RunningExecutionStatus,
		Timestamps: &shared.ExecutionTimestamps{
			RunningAt: runningAt,
		},
	}

	return serializeExecStateAndLogFailure(execState)
}

func serializedSuccess(runningAt *time.Time) string {
	now := time.Now()
	execState := &shared.ExecutionState{
		Status: shared.SucceededExecutionStatus,
		Timestamps: &shared.ExecutionTimestamps{
			RunningAt:  runningAt,
			FinishedAt: &now,
		},
	}

	return serializeExecStateAndLogFailure(execState)
}

func updateOnFailure(
	ctx context.Context,
	outputs string,
	msg string,
	condaPath string,
	runningAt *time.Time,
	integrationID uuid.UUID,
	integrationRepo repos.Integration,
	DB database.Database,
) {
	_, err := integrationRepo.Update(
		ctx,
		integrationID,
		map[string]interface{}{
			models.IntegrationConfig: (*collection_utils.Config)(&map[string]string{
				CondaPathKey: condaPath,
				ExecStateKey: serializedFailure(outputs, msg, runningAt),
			}),
		},
		DB,
	)
	if err != nil {
		log.Errorf("Failed to update conda integration: %v", err)
	}
}

func ValidateCondaDevelop() error {
	// This is to ensure we can use `conda develop` to update the python path later on.
	args := []string{
		"develop",
		"--help",
	}
	_, _, err := lib_utils.RunCmd(CondaCmdPrefix, args...)
	return err
}

func InitializeConda(
	ctx context.Context,
	integrationID uuid.UUID,
	integrationRepo repos.Integration,
	DB database.Database,
) {
	now := time.Now()
	_, err := integrationRepo.Update(
		ctx,
		integrationID,
		map[string]interface{}{
			models.IntegrationConfig: (*collection_utils.Config)(&map[string]string{
				ExecStateKey: serializedRunning(&now),
			}),
		},
		DB,
	)
	if err != nil {
		log.Errorf("Failed to update conda integration: %v", err)
		return
	}

	out, _, err := lib_utils.RunCmd(CondaCmdPrefix, "info", "--base")
	if err != nil {
		updateOnFailure(
			ctx,
			out,
			err.Error(),
			"", /* condaPath */
			&now,
			integrationID,
			integrationRepo,
			DB,
		)

		return
	}

	condaPath := strings.TrimSpace(out)

	err = createBaseEnvs()
	if err != nil {
		updateOnFailure(
			ctx,
			out,
			err.Error(),
			condaPath,
			&now,
			integrationID,
			integrationRepo,
			DB,
		)

		return
	}

	_, err = integrationRepo.Update(
		ctx,
		integrationID,
		map[string]interface{}{
			models.IntegrationConfig: (*collection_utils.Config)(&map[string]string{
				CondaPathKey: condaPath,
				ExecStateKey: serializedSuccess(&now),
			}),
		},
		DB,
	)

	if err != nil {
		log.Errorf("Failed to update conda integration: ")
	}
}

func GetCondaIntegration(
	ctx context.Context,
	userID uuid.UUID,
	integrationRepo repos.Integration,
	DB database.Database,
) (*models.Integration, error) {
	integrations, err := integrationRepo.GetByServiceAndUser(
		ctx,
		integration.Conda,
		userID,
		DB,
	)
	if err != nil {
		return nil, err
	}

	if len(integrations) == 0 {
		return nil, nil
	}

	return &integrations[0], nil
}

// ExtractConnectionState retrieves the current connection state from
// the given integration object.
// For non-conda integration, we assume they are always successfully connected
// since they are created in-sync in `connectIntegration` handler.
func ExtractConnectionState(
	integrationObject *models.Integration,
) (*shared.ExecutionState, error) {
	if integrationObject.Service != integration.Conda {
		return &shared.ExecutionState{
			Status: shared.SucceededExecutionStatus,
		}, nil
	}

	stateSerialized, ok := integrationObject.Config[ExecStateKey]
	if !ok {
		return &shared.ExecutionState{
			Status: shared.PendingExecutionStatus,
		}, nil
	}

	var state shared.ExecutionState
	err := json.Unmarshal([]byte(stateSerialized), &state)
	if err != nil {
		return nil, err
	}

	return &state, nil
}
