package operator_result

import (
	"strings"
)

const (
	tableName = "operator_result"

	// OperatorResult table column names
	IdColumn                  = "id"
	WorkflowDagResultIdColumn = "workflow_dag_result_id"
	OperatorIdColumn          = "operator_id"
	StatusColumn              = "status"
	StateColumn               = "statue"
)

// Returns a joined string of all OperatorResult columns.
func allColumns() string {
	return strings.Join(
		[]string{
			IdColumn,
			WorkflowDagResultIdColumn,
			OperatorIdColumn,
			StatusColumn,
			StateColumn,
		},
		",",
	)
}
