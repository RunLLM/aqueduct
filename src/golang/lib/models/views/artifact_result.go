package views

import (
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	mdl_shared "github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

// ArtifactResultStatus contains the status of an ArtifactResult as well
// as additional metadata
type ArtifactResultStatus struct {
	ArtifactID       uuid.UUID                  `db:"artifact_id" json:"artifact_id"`
	ArtifactResultID uuid.UUID                  `db:"artifact_result_id" json:"artifact_result_id"`
	DAGResultID      uuid.UUID                  `db:"workflow_dag_result_id" json:"workflow_dag_result_id"`
	Status           mdl_shared.ExecutionStatus `db:"status" json:"status"`
	Timestamp        time.Time                  `db:"timestamp" json:"timestamp"`
	ContentPath      string                     `db:"content_path" json:"content_path"`
}

// ArtifactWithResult is a concated view of artifact and artifact_result
// The only additional field is StorageConfig, which we fetch from dag.
// Storage is commonly used when fetching contents.
type ArtifactWithResult struct {
	ID          uuid.UUID               `db:"id" json:"id"`
	Name        string                  `db:"name" json:"name"`
	Description string                  `db:"description" json:"description"`
	Type        mdl_shared.ArtifactType `db:"type" json:"type"`

	ResultID    uuid.UUID                    `db:"result_id" json:"artifact_id"`
	DAGResultID uuid.UUID                    `db:"dag_result_id" json:"dag_result_id"`
	ContentPath string                       `db:"content_path" json:"content_path"`
	ExecState   shared.NullExecutionState    `db:"execution_state" json:"execution_state"`
	Metadata    artifact_result.NullMetadata `db:"metadata" json:"metadata"`

	StorageConfig shared.StorageConfig `db:"storage_config" json:"storage_config"`
}
