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

func UpdateOnSuccess(
	ctx context.Context,
	resourceType string,
	resourceConfig *shared.ResourceConfig,
	runningAt *time.Time,
	resourceID uuid.UUID,
	resourceRepo repos.Resource,
	DB database.Database,
) error {
	resourceConfigMap := (map[string]string)(*resourceConfig)
	resourceConfigMap[ExecStateKey] = SerializedSuccess(runningAt)
	updatedResourceConfig := (*shared.ResourceConfig)(&resourceConfigMap)

	_, err := resourceRepo.Update(
		ctx,
		resourceID,
		map[string]interface{}{
			models.ResourceConfig: updatedResourceConfig,
		},
		DB,
	)
	if err != nil {
		log.Errorf("Failed to update %s resource: %v", resourceType, err)
		return err
	}
	return nil
}

func UpdateOnFailure(
	ctx context.Context,
	outputs string,
	msg string,
	resourceType string,
	resourceConfig *shared.ResourceConfig,
	runningAt *time.Time,
	resourceID uuid.UUID,
	resourceRepo repos.Resource,
	DB database.Database,
) {
	resourceConfigMap := (map[string]string)(*resourceConfig)
	resourceConfigMap[ExecStateKey] = SerializedFailure(outputs, msg, runningAt)
	updatedResourceConfig := (*shared.ResourceConfig)(&resourceConfigMap)

	_, err := resourceRepo.Update(
		ctx,
		resourceID,
		map[string]interface{}{
			models.ResourceConfig: updatedResourceConfig,
		},
		DB,
	)
	if err != nil {
		log.Errorf("Failed to update %s resource: %v", resourceType, err)
	}
}

// ExtractConnectionState retrieves the current connection state from
// the given resource object. If the execution state is not present,
// we can assume that this is a legacy entry from before we always wrote
// an execution state. In this case, we always return success.
func ExtractConnectionState(
	resourceObject *models.Resource,
) (*shared.ExecutionState, error) {
	stateSerialized, ok := resourceObject.Config[ExecStateKey]
	if !ok {
		// TODO(ENG-2813): Remove this success logic. An execution state key should always exist.
		return &shared.ExecutionState{
			Status: shared.SucceededExecutionStatus,
		}, nil
	}

	var state shared.ExecutionState
	err := json.Unmarshal([]byte(stateSerialized), &state)
	if err != nil {
		return nil, err
	}

	return &state, nil
}
