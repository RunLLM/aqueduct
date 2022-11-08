package models

import (
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/google/uuid"
)

const (
	ArtifactResultTable = "artifact_result"

	// ArtifactResult table column names
	ArtifactResultID                  = "id"
	ArtifactResultWorkflowDAGResultID = "workflow_dag_result_id"
	ArtifactResultArtifactID         = "artifact_id"
	ArtifactResultContentPath         = "content_path"

	// `Status` is initialized to "PENDING" for each new artifact result.
	ArtifactResultStatus   = "status"
	ArtifactResultMetadata = "metadata"

	// `ExecState` is initialized to nil. Expected to be set on updates only.
	ArtifactResultExecState = "execution_state"
)

// An ArtifactResult maps to the artifact_result table.
type ArtifactResult struct {
	ID              uuid.UUID              `db:"id" json:"id"`
	WorkflowDagResultId          uuid.UUID              `db:"workflow_dag_result_id" json:"workflow_dag_result_id"`
	ArtifactId          uuid.UUID `db:"artifact_id" json:"artifact_id"`
	ContentPath            string                 `db:"content_path" json:"content_path"`
	ExecState shared.NullExecutionState `db:"execution_state" json:"execution_state"`
	Metadata  utils.NullMetadata              `db:"metadata" json:"metadata"`
}

// ArtifactResultCols returns a comma-separated string of all ArtifactResult columns.
func ArtifactResultCols() string {
	return strings.Join(allArtifactResultCols(), ",")
}

// ArtifactResultColsWithPrefix returns a comma-separated string of all
// ArtifactResult columns prefixed by the table name.
func ArtifactResultColsWithPrefix() string {
	cols := allArtifactResultCols()
	for i, col := range cols {
		cols[i] = fmt.Sprintf("%s.%s", ArtifactResultTable, col)
	}

	return strings.Join(cols, ",")
}

func allArtifactResultCols() []string {
	return []string{
		ArtifactResultID,
		ArtifactResultWorkflowDAGResultID,
		ArtifactResultArtifactID,
		ArtifactResultContentPath,
		ArtifactResultStatus,
		ArtifactResultMetadata,
		ArtifactResultExecState,
	}
}