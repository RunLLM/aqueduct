package shared

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/dropbox/godropbox/errors"
)

const (
	githubIssueLink    = "https://github.com/aqueducthq/aqueduct/issues/new?assignees=&labels=bug&template=bug_report.md&title=%5BBUG%5D"
	TipCreateBugReport = "We are sorry to see this :(. " +
		"You could send over a bug report through github issue: " +
		githubIssueLink +
		" , or in our slack channel. We will get back to you as soon as we can."
	TipUnknownInternalError = "An unexpected error occurred. " + TipCreateBugReport
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

type Logs struct {
	Stdout string `json:"stdout"`
	StdErr string `json:"stderr"`
}

type Error struct {
	Context string `json:"context"`
	Tip     string `json:"tip"`
}

type ExecutionLogs struct {
	UserLogs      *Logs           `json:"user_logs"`
	Code          ExecutionStatus `json:"code"`
	FailureReason FailureType     `json:"failure_reason"`
	Error         *Error          `json:"error"`
}

func (e *ExecutionLogs) Value() (driver.Value, error) {
	return utils.ValueJsonB(*e)
}

func (e *ExecutionLogs) Scan(value interface{}) error {
	return utils.ScanJsonB(value, e)
}
