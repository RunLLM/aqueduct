package shared

import (
	"github.com/dropbox/godropbox/errors"
)

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