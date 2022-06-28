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

type FailureType int64

const (
	SystemFailure FailureType = 0
	UserFailure   FailureType = 1
	NoFailure     FailureType = 2
)
