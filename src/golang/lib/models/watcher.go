package models

import (
	"strings"

	"github.com/google/uuid"
)

const (
	WatcherTable = "workflow_watcher"

	// Watcher column names
	WatcherWorkflowID = "workflow_id"
	WatcherUserID     = "user_id"
)

// A Watcher maps to the workflow_watcher table.
type Watcher struct {
	WorkflowID uuid.UUID `db:"workflow_id"`
	UserID     uuid.UUID `db:"user_id"`
}

// WatcherCols returns a comma-separated string of all Watcher columns.
func WatcherCols() string {
	return strings.Join(allWatcherCols(), ",")
}

func allWatcherCols() []string {
	return []string{
		WatcherWorkflowID,
		WatcherUserID,
	}
}
