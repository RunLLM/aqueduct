package shared

import (
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
	// 'erased' refers to 'erased after success'.
	// Caller should consider 'erased' as success for error handling,
	// but should not expect any non-metadata content to be available.
	ErasedExecutionStatus  ExecutionStatus = "erased"
	UnknownExecutionStatus ExecutionStatus = "unknown"
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
