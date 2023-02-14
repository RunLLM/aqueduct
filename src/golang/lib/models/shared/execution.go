package shared

import (
	"database/sql/driver"
	"time"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/dropbox/godropbox/errors"
)

// This should mirror all ExecutionStatus

type ExecutionTimestamps struct {
	RegisteredAt *time.Time `json:"registered_at"`
	PendingAt    *time.Time `json:"pending_at"`
	RunningAt    *time.Time `json:"running_at"`
	FinishedAt   *time.Time `json:"finished_at"`
}

type ExecutionState struct {
	UserLogs *Logs           `json:"user_logs"`
	Status   ExecutionStatus `json:"status"`

	// These fields are only set if status == Failed.
	FailureType *FailureType         `json:"failure_type"`
	Error       *Error               `json:"error"`
	Timestamps  *ExecutionTimestamps `json:"timestamps"`
}

func (e ExecutionState) Terminated() bool {
	return e.Status == FailedExecutionStatus || e.Status == SucceededExecutionStatus || e.Status == CanceledExecutionStatus
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

	logs := &ExecutionState{}
	if err := logs.Scan(value); err != nil {
		return err
	}

	n.ExecutionState, n.IsNull = *logs, false
	return nil
}

type ExecutionStatus string

const (
	// Registered is a special state that indicates a object has been registered
	// but has no runs yet. This is typically used in workflows.
	RegisteredExecutionStatus ExecutionStatus = "registered"
	PendingExecutionStatus    ExecutionStatus = "pending"
	RunningExecutionStatus    ExecutionStatus = "running"
	CanceledExecutionStatus   ExecutionStatus = "canceled"
	FailedExecutionStatus     ExecutionStatus = "failed"
	SucceededExecutionStatus  ExecutionStatus = "succeeded"
	UnknownExecutionStatus    ExecutionStatus = "unknown"
)

type NullExecutionStatus struct {
	ExecutionStatus
	IsNull bool
}

func (n *NullExecutionStatus) Scan(value interface{}) error {
	if value == nil {
		n.IsNull = true
		return nil
	}

	s, ok := value.(string)
	if !ok {
		return errors.New("Type assertion to string failed")
	}

	n.ExecutionStatus, n.IsNull = ExecutionStatus(s), false
	return nil
}
