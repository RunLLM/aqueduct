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
			Stdout: outputs,
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

func updateFailure(
	ctx context.Context,
	outputs string,
	msg string,
	condaPath string,
	runningAt *time.Time,
	integrationID uuid.UUID,
	integrationWriter integration.Writer,
	db database.Database,
) {
	_, err := integrationWriter.UpdateIntegration(
		ctx,
		integrationID,
		map[string]interface{}{
			"config": (*collection_utils.Config)(&map[string]string{
				CondaPathKey: condaPath,
				ExecStateKey: serializedFailure(outputs, msg, runningAt),
			}),
		},
		db,
	)
	if err != nil {
		log.Errorf("Failed to update conda integration: %v", err)
	}
}

func InitializeConda(
	ctx context.Context,
	integrationID uuid.UUID,
	integrationWriter integration.Writer,
	db database.Database,
) {
	now := time.Now()
	_, err := integrationWriter.UpdateIntegration(
		ctx,
		integrationID,
		map[string]interface{}{
			"config": (*collection_utils.Config)(&map[string]string{
				ExecStateKey: serializedRunning(&now),
			}),
		},
		db,
	)
	if err != nil {
		log.Errorf("Failed to update conda integration: %v", err)
		return
	}

	out, _, err := lib_utils.RunCmd(CondaCmdPrefix, "info", "--base")
	if err != nil {
		updateFailure(
			ctx,
			out,
			err.Error(),
			"", /* condaPath */
			&now,
			integrationID,
			integrationWriter,
			db,
		)

		return
	}

	condaPath := strings.TrimSpace(out)
	log.Infof("Conda base path is %s", condaPath)

	err = createBaseEnvs(condaPath)
	if err != nil {
		updateFailure(
			ctx,
			out,
			err.Error(),
			condaPath,
			&now,
			integrationID,
			integrationWriter,
			db,
		)

		return
	}

	// This is to ensure we can use `conda develop` to update the python path later on.
	args := []string{
		"install",
		"conda-build",
		"-y",
	}
	_, _, err = lib_utils.RunCmd(CondaCmdPrefix, args...)
	if err != nil {
		updateFailure(
			ctx,
			out,
			err.Error(),
			condaPath,
			&now,
			integrationID,
			integrationWriter,
			db,
		)

		return
	}

	_, err = integrationWriter.UpdateIntegration(
		ctx,
		integrationID,
		map[string]interface{}{
			"config": (*collection_utils.Config)(&map[string]string{
				CondaPathKey: condaPath,
				ExecStateKey: serializedSuccess(&now),
			}),
		},
		db,
	)

	if err != nil {
		log.Errorf("Failed to update conda integration: ")
	}
}

func getCondaIntegration(
	ctx context.Context,
	userId uuid.UUID,
	integrationReader integration.Reader,
	db database.Database,
) (*integration.Integration, error) {
	integrations, err := integrationReader.GetIntegrationsByServiceAndUser(
		ctx,
		integration.Conda,
		userId,
		db,
	)
	if err != nil {
		return nil, err
	}

	if len(integrations) == 0 {
		return nil, nil
	}

	return &integrations[0], nil
}

// CondaConnectionState shows the current conda connection status
// if the user is connected to conda. Otherwise, it returns nil.
func CondaConnectionState(
	ctx context.Context,
	userId uuid.UUID,
	integrationReader integration.Reader,
	db database.Database,
) (*integration.Integration, *shared.ExecutionState, error) {
	integration, err := getCondaIntegration(ctx, userId, integrationReader, db)
	if err != nil {
		return nil, nil, err
	}

	if integration == nil {
		return nil, nil, nil
	}

	stateSerialized, ok := integration.Config[ExecStateKey]
	if !ok {
		return integration, &shared.ExecutionState{
			Status: shared.PendingExecutionStatus,
		}, nil
	}

	var state shared.ExecutionState
	err = json.Unmarshal([]byte(stateSerialized), &state)
	if err != nil {
		return nil, nil, err
	}

	return integration, &state, nil
}
