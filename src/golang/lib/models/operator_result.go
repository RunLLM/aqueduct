package models

import (
	"strings"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

const (
	OperatorResultTable = "operator_result"

	// OperatorResult table column names
	OperatorResultID          = "id"
	OperatorResultDAGResultID = "workflow_dag_result_id"
	OperatorResultOperatorID  = "operator_id"

	// `Status` is initialized to "PENDING" for each new operator result.
	OperatorResultStatus = "status"

	// `ExecState` is initialized to nil. Expected to be set on updates only.
	OperatorResultExecState = "execution_state"
)

// A OperatorResult maps to the operator_result table.
type OperatorResult struct {
	ID          uuid.UUID `db:"id" json:"id"`
	DAGResultID uuid.UUID `db:"workflow_dag_result_id" json:"workflow_dag_result_id"`
	OperatorID  uuid.UUID `db:"operator_id" json:"operator_id"`
	// TODO(ENG-1453): Remove status. This field is redundant now that ExecState exists.
	//  Avoid using status in new code.
	Status    shared.ExecutionStatus    `db:"status" json:"status"`
	ExecState shared.NullExecutionState `db:"execution_state" json:"execution_state"`
}

// OperatorResultCols returns a comma-separated string of all OperatorResult columns.
func OperatorResultCols() string {
	return strings.Join(allOperatorResultCols(), ",")
}

func allOperatorResultCols() []string {
	return []string{
		OperatorResultID,
		OperatorResultDAGResultID,
		OperatorResultOperatorID,
		OperatorResultStatus,
		OperatorResultExecState,
	}
}
