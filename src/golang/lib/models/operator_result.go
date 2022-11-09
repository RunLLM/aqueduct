package models

import (
	"time"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

const (
	OperatorResultTable = "operator_result"

	// OperatorResult table column names
	OperatorResultID                  = "id"
	OperatorResultWorkflowDAGResultID = "workflow_dag_result_id"
	OperatorResultOperatorID          = "operator_id"

	// `Status` is initialized to "PENDING" for each new operator result.
	OperatorResultStatus = "status"

	// `ExecState` is initialized to nil. Expected to be set on updates only.
	OperatorResultExecState = "execution_state"
)

// A OperatorResult maps to the operator_result table.
type OperatorResult struct {
	Id                  uuid.UUID `db:"id" json:"id"`
	WorkflowDagResultID uuid.UUID `db:"workflow_dag_result_id" json:"workflow_dag_result_id"`
	OperatorIdD         uuid.UUID `db:"operator_id" json:"operator_id"`
	ExecState shared.NullExecutionState `db:"execution_state" json:"execution_state"`
}

// OperatorResultCols returns a comma-separated string of all OperatorResult columns.
func OperatorResultCols() string {
	return strings.Join(allOperatorResultCols(), ",")
}

func allOperatorResultCols() []string {
	return []string{
		OperatorResultID,
		OperatorResultWorkflowDAGResultID,
		OperatorResultOperatorID,
		OperatorResultStatus,
		OperatorResultExecState,
	}
}