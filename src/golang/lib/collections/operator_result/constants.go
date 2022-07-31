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

	// `Status` is initialized to "PENDING" for each new operator result.
	StatusColumn = "status"

	// `ExecState` is initialized to nil. Expected to be set on updates only.
	ExecStateColumn = "execution_state"
)

// Returns a joined string of all OperatorResult columns.
func allColumns() string {
	return strings.Join(
		[]string{
			IdColumn,
			WorkflowDagResultIdColumn,
			OperatorIdColumn,
			StatusColumn,
			ExecStateColumn,
		},
		",",
	)
}
