package views

import (
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
