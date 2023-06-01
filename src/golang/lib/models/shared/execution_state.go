package shared

import (
	"database/sql/driver"
	"time"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
)

type ExecutionState struct {
	UserLogs *Logs           `json:"user_logs"`
	Status   ExecutionStatus `json:"status"`

	// These two failure fields are only set if status == Failed.
	FailureType *FailureType `json:"failure_type"`
	Error       *Error       `json:"error"`

	Timestamps *ExecutionTimestamps `json:"timestamps"`
}

func (e ExecutionState) Terminated() bool {
	return e.Status == FailedExecutionStatus || e.Status == SucceededExecutionStatus || e.Status == CanceledExecutionStatus || e.Status == ErasedExecutionStatus
}

func (e *ExecutionState) HasBlockingFailure() bool {
	return e.Status == FailedExecutionStatus && *e.FailureType != UserNonFatalFailure
}

func (e *ExecutionState) HasWarning() bool {
	return e.Status == FailedExecutionStatus && !e.HasBlockingFailure()
}

func (e *ExecutionState) HasSystemError() bool {
	return e.Status == FailedExecutionStatus && *e.FailureType == SystemFailure
}

// UpdateWithFailure also updates the `FinishedAt` timestamp.
func (e *ExecutionState) UpdateWithFailure(failureType FailureType, execErr *Error) {
	e.Status = FailedExecutionStatus
	e.FailureType = &failureType
	e.Error = execErr

	finishedAt := time.Now()
	e.Timestamps.FinishedAt = &finishedAt
}

func (e *ExecutionState) Value() (driver.Value, error) {
	return utils.ValueJSONB(*e)
}

func (e *ExecutionState) Scan(value interface{}) error {
	return utils.ScanJSONB(value, e)
}

type NullExecutionState struct {
	ExecutionState
	IsNull bool
}

func (n *NullExecutionState) Value() (driver.Value, error) {
	if n.IsNull {
		return nil, nil
	}

	return (&n.ExecutionState).Value()
}

func (n *NullExecutionState) Scan(value interface{}) error {
	if value == nil {
		n.IsNull = true
		return nil
	}

	executionState := &ExecutionState{}
	if err := executionState.Scan(value); err != nil {
		return err
	}

	n.ExecutionState, n.IsNull = *executionState, false
	return nil
}
