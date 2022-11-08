package shared

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
)

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
