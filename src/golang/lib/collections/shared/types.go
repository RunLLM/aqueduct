package shared

import (
	"github.com/dropbox/godropbox/errors"
)

var ErrInvalidStorageConfig = errors.New("Invalid Storage Config")

type ExecutionStatus string

const (
	SucceededExecutionStatus ExecutionStatus = "succeeded"
	FailedExecutionStatus    ExecutionStatus = "failed"
	PendingExecutionStatus   ExecutionStatus = "pending"
	UnknownExecutionStatus   ExecutionStatus = "unknown"
)
