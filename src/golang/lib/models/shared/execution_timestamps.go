package shared

import (
	"time"

	"github.com/dropbox/godropbox/errors"
)

type ExecutionTimestamps struct {
	RegisteredAt *time.Time `json:"registered_at"`
	PendingAt    *time.Time `json:"pending_at"`
	RunningAt    *time.Time `json:"running_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

// ExecutionTimestampsJsonFieldByStatus returns the json_field
// of the timestamp for a given status. This is useful when
// we are directly inserting the timestamp to DB's JSON blob
// based on the status, using a `json_set()` query.
func ExecutionTimestampsJsonFieldByStatus(
	status ExecutionStatus,
) (string, error) {
	if status == SucceededExecutionStatus ||
		status == FailedExecutionStatus ||
		status == CanceledExecutionStatus {
		return "finished_at", nil
	}

	if status == RunningExecutionStatus {
		return "running_at", nil
	}

	if status == PendingExecutionStatus {
		return "pending_at", nil
	}

	if status == RegisteredExecutionStatus {
		return "registered_at", nil
	}

	if status == DeletedExecutionStatus {
		return "deleted_at", nil
	}

	return "", errors.Newf("Execution status %s is not valid in timestamps", status)
}
