package shared

import (
	"time"
)

// This should mirror all ExecutionStatus

type ExecutionTimestamps struct {
	RegisteredAt *time.Time `json:"registered_at"`
	PendingAt    *time.Time `json:"pending_at"`
	RunningAt    *time.Time `json:"running_at"`
	FinishedAt   *time.Time `json:"finished_at"`
}
