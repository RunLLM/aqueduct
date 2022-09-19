package workflow_dag_result

import (
	"fmt"
	"strings"
)

const (
	tableName = "workflow_dag_result"

	// WorkflowDagResult table column names
	IdColumn            = "id"
	WorkflowDagIdColumn = "workflow_dag_id"
	StatusColumn        = "status"
	CreatedAtColumn     = "created_at"
	ExecStateColumn     = "execution_state"
)

// Returns a joined string of all WorkflowDagResult columns.
func allColumns() string {
	return strings.Join(
		[]string{
			IdColumn,
			WorkflowDagIdColumn,
			StatusColumn,
			CreatedAtColumn,
			ExecStateColumn,
		},
		",",
	)
}

// Returns a joined string of all WorkflowDagResult columns prefixed by table name.
func allColumnsWithPrefix() string {
	return strings.Join(
		[]string{
			fmt.Sprintf("%s.%s", tableName, IdColumn),
			fmt.Sprintf("%s.%s", tableName, WorkflowDagIdColumn),
			fmt.Sprintf("%s.%s", tableName, StatusColumn),
			fmt.Sprintf("%s.%s", tableName, CreatedAtColumn),
		},
		",",
	)
}
