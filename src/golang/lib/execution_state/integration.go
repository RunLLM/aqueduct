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
		log.Errorf("Failed to update conda integration: %v", err)
	}
}

// ExtractConnectionState retrieves the current connection state from
// the given integration object.
// For non-conda integration, we assume they are always successfully connected
// since they are created in-sync in `connectIntegration` handler.
func ExtractConnectionState(
	integrationObject *models.Integration,
) (*shared.ExecutionState, error) {
	if integrationObject.Service != shared.Conda {
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
