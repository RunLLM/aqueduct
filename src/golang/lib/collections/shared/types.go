package shared

import (
	"database/sql/driver"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/dropbox/godropbox/errors"
)

const (
	githubIssueLink    = "https://github.com/aqueducthq/aqueduct/issues/new?assignees=&labels=bug&template=bug_report.md&title=%5BBUG%5D"
	TipCreateBugReport = "Please create bug report in github: " +
		githubIssueLink + " . " +
		"We will get back to you as soon as we can."
	TipUnknownInternalError = "Sorry, we've run into an unexpected error! " + TipCreateBugReport
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

type NotificationLevel string

const (
	SuccessNotificationLevel NotificationLevel = "success"
	WarningNotificationLevel NotificationLevel = "warning"
	ErrorNotificationLevel   NotificationLevel = "error"
)

type FailureType int64

const (
	Success          FailureType = 0
	SystemFailure    FailureType = 1
	UserFatalFailure FailureType = 2

	// Orchestration can continue onwards, despite this failure.
	// Eg. Check operator with WARNING severity does not pass.
	UserNonFatalFailure FailureType = 3
)

type Logs struct {
	Stdout string `json:"stdout"`
	StdErr string `json:"stderr"`
}

type Error struct {
	Context string `json:"context"`
	Tip     string `json:"tip"`
}

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

func (e *ExecutionState) Value() (driver.Value, error) {
	return utils.ValueJsonB(*e)
}

func (e *ExecutionState) Scan(value interface{}) error {
	return utils.ScanJsonB(value, e)
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
