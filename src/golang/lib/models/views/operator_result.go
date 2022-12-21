package views

import (
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/google/uuid"
)

// OperatorResultStatus is a wrapper around the ExecutionState of an
// OperatorResult and additional metadata.
type OperatorResultStatus struct {
	ArtifactID   uuid.UUID              `db:"artifact_id"`
	Metadata     *shared.ExecutionState `db:"metadata"`
	DAGResultID  uuid.UUID              `db:"workflow_dag_result_id"`
	OperatorName utils.NullString       `db:"operator_name"`
}

// OperatorWithResult is a concated view of operator and operator_result
type OperatorWithResult struct {
	ID                     uuid.UUID      `db:"id" json:"id"`
	Name                   string         `db:"name" json:"name"`
	Description            string         `db:"description" json:"description"`
	Spec                   operator.Spec  `db:"spec" json:"spec"`
	ExecutionEnvironmentID utils.NullUUID `db:"execution_environment_id" json:"execution_environment_id"`

	ResultID    uuid.UUID `db:"result_id" json:"id"`
	DAGResultID uuid.UUID `db:"dag_result_id" json:"dag_result_id"`

	// TODO(ENG-1453): Remove status. This field is redundant now that ExecState exists.
	//  Avoid using status in new code.
	Status    shared.ExecutionStatus    `db:"status" json:"status"`
	ExecState shared.NullExecutionState `db:"execution_state" json:"execution_state"`
}
