package workflow_watcher

import "strings"

const (
	tableName = "workflow_watcher"

	// WorkflowWatcher table column names
	WorkflowIdColumn = "workflow_id"
	UserIdColumn     = "user_id"
)

// Returns a joined string of all WorkflowWatcher columns.
func allColumns() string {
	return strings.Join(
		[]string{
			WorkflowIdColumn,
			UserIdColumn,
		},
		",",
	)
}
