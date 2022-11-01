package models

import (
	"time"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/google/uuid"
)

// An ArtifactResult maps to the artifact_result table.
type ArtifactResult struct {
	ID              uuid.UUID              `db:"id" json:"id"`
	WorkflowDagResultId          uuid.UUID              `db:"workflow_dag_result_id" json:"workflow_dag_result_id"`
	ArtifactId          uuid.UUID `db:"artifact_id" json:"artifact_id"`
	ContentPath            string                 `db:"content_path" json:"content_path"`
	ExecState shared.NullExecutionState `db:"execution_state" json:"execution_state"`
	Metadata  primitive.NullMetadata              `db:"metadata" json:"metadata"`
}
