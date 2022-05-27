package workflow

import (
	"fmt"
	"strings"
)

const (
	tableName = "workflow"

	// Workflow table column names
	IdColumn          = "id"
	UserIdColumn      = "user_id"
	NameColumn        = "name"
	DescriptionColumn = "description"
	ScheduleColumn    = "schedule"
	CreatedAtColumn   = "created_at"
	RetentionColumn   = "retention_policy"
)

// Returns a joined string of all Workflow columns.
func allColumns() string {
	return strings.Join(
		[]string{
			IdColumn,
			UserIdColumn,
			NameColumn,
			DescriptionColumn,
			ScheduleColumn,
			CreatedAtColumn,
			RetentionColumn,
		},
		",",
	)
}

// Returns a joined string of all Workflow columns prefixed by table name.
func allColumnsWithPrefix() string {
	return strings.Join(
		[]string{
			fmt.Sprintf("%s.%s", tableName, IdColumn),
			fmt.Sprintf("%s.%s", tableName, UserIdColumn),
			fmt.Sprintf("%s.%s", tableName, NameColumn),
			fmt.Sprintf("%s.%s", tableName, DescriptionColumn),
			fmt.Sprintf("%s.%s", tableName, ScheduleColumn),
			fmt.Sprintf("%s.%s", tableName, CreatedAtColumn),
			fmt.Sprintf("%s.%s", tableName, RetentionColumn),
		},
		",",
	)
}
