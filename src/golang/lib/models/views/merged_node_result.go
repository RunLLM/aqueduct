package views

import (
	"fmt"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

const (
	OperatorWithArtifactNodeResultTable = "merged_node_result"

	// OperatorWithArtifactNodeResult table column names
	OperatorWithArtifactNodeResultID                = "id"
	OperatorWithArtifactNodeResultOperatorExecState = "operator_exec_state"
	OperatorWithArtifactNodeResultArtifactID        = "artifact_id"
	OperatorWithArtifactNodeResultMetadata          = "metadata"
	OperatorWithArtifactNodeResultContentPath       = "content_path"
	OperatorWithArtifactNodeResultArtifactExecState = "artifact_exec_state"
)

// An OperatorWithArtifactNodeResult maps to the merged_node_result table.
type OperatorWithArtifactNodeResult struct {
	ID                uuid.UUID                         `db:"id" json:"id"`
	OperatorExecState shared.NullExecutionState         `db:"operator_exec_state" json:"operator_exec_state"`
	ArtifactID        uuid.UUID                         `db:"artifact_id" json:"artifact_id"`
	Metadata          shared.NullArtifactResultMetadata `db:"metadata" json:"metadata"`
	ContentPath       string                            `db:"content_path" json:"content_path"`
	ArtifactExecState shared.NullExecutionState         `db:"artifact_exec_state" json:"artifact_exec_state"`
}

// OperatorWithArtifactNodeResultCols returns a comma-separated string of all OperatorWithArtifactNodeResult columns.
func OperatorWithArtifactNodeResultCols() string {
	return strings.Join(allOperatorWithArtifactNodeResultCols(), ",")
}

// OperatorWithArtifactNodeResultColsWithPrefix returns a comma-separated string of all
// OperatorWithArtifactNodeResult columns prefixed by the table name.
func OperatorWithArtifactNodeResultColsWithPrefix() string {
	cols := allOperatorWithArtifactNodeResultCols()
	for i, col := range cols {
		cols[i] = fmt.Sprintf("%s.%s", OperatorWithArtifactNodeResultTable, col)
	}

	return strings.Join(cols, ",")
}

func allOperatorWithArtifactNodeResultCols() []string {
	return []string{
		OperatorWithArtifactNodeResultID,
		OperatorWithArtifactNodeResultOperatorExecState,
		OperatorWithArtifactNodeResultArtifactID,
		OperatorWithArtifactNodeResultMetadata,
		OperatorWithArtifactNodeResultContentPath,
		OperatorWithArtifactNodeResultArtifactExecState,
	}
}
