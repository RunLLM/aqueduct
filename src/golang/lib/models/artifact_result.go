package models

import (
	"fmt"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

const (
	ArtifactResultTable = "artifact_result"

	// ArtifactResult table column names
	ArtifactResultID          = "id"
	ArtifactResultDAGResultID = "workflow_dag_result_id"
	ArtifactResultArtifactID  = "artifact_id"
	ArtifactResultContentPath = "content_path"

	// `Status` is initialized to "PENDING" for each new artifact result.
	// TODO(ENG-1453): Remove status. This field is redundant now that ExecState exists.
	// Avoid using status in new code.
	// Cannot remove it now because CREATE won't work since Status has a NON-NULL constraint.
	ArtifactResultStatus   = "status"
	ArtifactResultMetadata = "metadata"

	// `ExecState` is initialized to nil. Expected to be set on updates only.
	ArtifactResultExecState = "execution_state"
)

// An ArtifactResult maps to the artifact_result table.
type ArtifactResult struct {
	ID          uuid.UUID                         `db:"id" json:"id"`
	DAGResultID uuid.UUID                         `db:"workflow_dag_result_id" json:"workflow_dag_result_id"`
	ArtifactID  uuid.UUID                         `db:"artifact_id" json:"artifact_id"`
	ContentPath string                            `db:"content_path" json:"content_path"`
	Status      shared.ExecutionStatus            `db:"status" json:"status"`
	ExecState   shared.NullExecutionState         `db:"execution_state" json:"execution_state"`
	Metadata    shared.NullArtifactResultMetadata `db:"metadata" json:"metadata"`
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
		ArtifactResultDAGResultID,
		ArtifactResultArtifactID,
		ArtifactResultContentPath,
		ArtifactResultStatus,
		ArtifactResultMetadata,
		ArtifactResultExecState,
	}
}
