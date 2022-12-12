package views

import "github.com/google/uuid"

// DAGResultWorkflowMetadata contains Workflow metadata for a DAGResult.
type DAGResultWorkflowMetadata struct {
	WorkflowID  uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	DAGResultID uuid.UUID `db:"dag_result_id" json:"dag_result_id"`
}
