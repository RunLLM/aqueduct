// Execution Loggings
package logging

import (
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
)

const (
	githubIssueLink    = "https://github.com/aqueducthq/aqueduct/issues/new?assignees=&labels=bug&template=bug_report.md&title=%5BBUG%5D"
	TipCreateBugReport = "We are sorry to see this :(. " +
		"You could send over a bug report through github issue: " +
		githubIssueLink +
		" , or in our slack channel. We will get back to you as soon as we can."
	TipUnknownInternalError = "An unexpected error occurred. " + TipCreateBugReport
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
	UserLogs      *Logs                  `json:"user_logs"`
	Code          shared.ExecutionStatus `json:"code"`
	FailureReason shared.FailureType     `json:"failure_reason"`
	Error         *Error                 `json:"error"`
}
