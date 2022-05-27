package workflow_dag_edge

import "strings"

const (
	tableName = "workflow_dag_edge"

	// WorkflowDagEdge table column names
	WorkflowDagIdColumn = "workflow_dag_id"
	TypeColumn          = "type"
	FromIdColumn        = "from_id"
	ToIdColumn          = "to_id"
	IdxColumn           = "idx"
)

// Returns a joined string of all WorkflowDagEdge columns.
func allColumns() string {
	return strings.Join(
		[]string{
			WorkflowDagIdColumn,
			TypeColumn,
			FromIdColumn,
			ToIdColumn,
			IdxColumn,
		},
		",",
	)
}
