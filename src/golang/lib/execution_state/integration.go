package execution_state

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	ExecStateKey = "exec_state"
)

func SerializeExecStateAndLogFailure(execState *shared.ExecutionState) string {
	serializedState, err := json.Marshal(execState)
	if err != nil {
		// We should never hit this
		log.Errorf("Error marshalling serialized state: %v", err)
		return ""
	}

	return string(serializedState)
}

func SerializedFailure(
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

	return SerializeExecStateAndLogFailure(execState)
}

func SerializedRunning(runningAt *time.Time) string {
	execState := &shared.ExecutionState{
		Status: shared.RunningExecutionStatus,
		Timestamps: &shared.ExecutionTimestamps{
			RunningAt: runningAt,
		},
	}

	return SerializeExecStateAndLogFailure(execState)
}

func SerializedSuccess(runningAt *time.Time) string {
	now := time.Now()
	execState := &shared.ExecutionState{
		Status: shared.SucceededExecutionStatus,
		Timestamps: &shared.ExecutionTimestamps{
			RunningAt:  runningAt,
			FinishedAt: &now,
		},
	}

	return SerializeExecStateAndLogFailure(execState)
}

func UpdateOnFailure(
	ctx context.Context,
	outputs string,
	msg string,
	integration_type string,
	integrationConfig *shared.IntegrationConfig,
	runningAt *time.Time,
	integrationID uuid.UUID,
	integrationRepo repos.Integration,
	DB database.Database,
) {
	integrationConfigMap := (map[string]string)(*integrationConfig)
	integrationConfigMap[ExecStateKey] = SerializedFailure(outputs, msg, runningAt)
	updatedIntegrationConfig := (*shared.IntegrationConfig)(&integrationConfigMap)

	_, err := integrationRepo.Update(
		ctx,
		integrationID,
		map[string]interface{}{
			models.IntegrationConfig: updatedIntegrationConfig,
		},
		DB,
	)
	if err != nil {
		log.Errorf("Failed to update %s integration: %v",integration_type ,err)
	}
}

// ExtractConnectionState retrieves the current connection state from
// the given integration object.
// For integrations other than lambda and conda, we assume they are always successfully connected
// since they are created in-sync in `connectIntegration` handler.
func ExtractConnectionState(
	integrationObject *models.Integration,
) (*shared.ExecutionState, error) {
	if integrationObject.Service != shared.Conda &&
		integrationObject.Service != shared.Lambda {
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
