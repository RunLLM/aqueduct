package views

import (
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator/connector"
	"github.com/google/uuid"
)

// LoadOperator contains metadata about a Load Operator
type LoadOperator struct {
	OperatorName    string    `db:"operator_name" json:"operator_name"`
	ModifiedAt      time.Time `db:"modified_at" json:"modified_at"`
	IntegrationName string    `db:"integration_name" json:"integration_name"`

	Spec *connector.Load `db:"spec" json:"spec"`
}

// LoadOperatorSpec is a wrapper around a Load Operator's spec and other metadata
type LoadOperatorSpec struct {
	ArtifactID   uuid.UUID     `db:"artifact_id" json:"artifact_id"`
	ArtifactName string        `db:"artifact_name" json:"artifact_name"`
	OperatorID   uuid.UUID     `db:"load_operator_id" json:"load_operator_id"`
	WorkflowName string        `db:"workflow_name" json:"workflow_name"`
	WorkflowID   uuid.UUID     `db:"workflow_id" json:"workflow_id"`
	Spec         operator.Spec `db:"spec" json:"spec"`
}

// OperatorRelation is a wrapper around an Operator's ID and the IDs of
// the Workflow and DAG it is associated to.
type OperatorRelation struct {
	WorkflowID uuid.UUID `db:"workflow_id" json:"workflow_id"`
	DagID      uuid.UUID `db:"workflow_dag_id" json:"workflow_dag_id"`
	OperatorID uuid.UUID `db:"operator_id" json:"operator_id"`
}
