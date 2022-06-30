package shared

import (
	"database/sql/driver"

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
	Success       FailureType = 0
	SystemFailure FailureType = 1
	UserFailure   FailureType = 2
)

type Logs struct {
	Stdout string `json:"stdout"`
	StdErr string `json:"stderr"`
}

type Error struct {
	Context string `json:"context"`
	Tip     string `json:"tip"`
}

type ExecutionState struct {
	UserLogs    *Logs           `json:"user_logs"`
	Status      ExecutionStatus `json:"status"`
	FailureType FailureType     `json:"failure_type"`
	Error       *Error          `json:"error"`
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
