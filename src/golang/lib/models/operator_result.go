package models

import (
	"time"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

// A OperatorResult maps to the operator_result table.
type OperatorResult struct {
	Id                  uuid.UUID `db:"id" json:"id"`
	WorkflowDagResultId uuid.UUID `db:"workflow_dag_result_id" json:"workflow_dag_result_id"`
	OperatorId          uuid.UUID `db:"operator_id" json:"operator_id"`
	ExecState shared.NullExecutionState `db:"execution_state" json:"execution_state"`
}