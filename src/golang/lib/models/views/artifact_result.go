package views

import (
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/google/uuid"
)

// ArtifactResultStatus contains the status of an ArtifactResult as well
// as additional metadata
type ArtifactResultStatus struct {
	ArtifactID       uuid.UUID              `db:"artifact_id" json:"artifact_id"`
	ArtifactResultID uuid.UUID              `db:"artifact_result_id" json:"artifact_result_id"`
	DAGResultID      uuid.UUID              `db:"workflow_dag_result_id" json:"workflow_dag_result_id"`
	Status           shared.ExecutionStatus `db:"status" json:"status"`
	Timestamp        time.Time              `db:"timestamp" json:"timestamp"`
	ContentPath      string                 `db:"content_path" json:"content_path"`
}
