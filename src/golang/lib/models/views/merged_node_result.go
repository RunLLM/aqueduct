package views

import (
	"fmt"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

const (
	MergedNodeResultTable = "merged_node_result"

	// MergedNodeResult table column names
	MergedNodeResultID                = "id"
	MergedNodeResultOperatorExecState = "operator_exec_state"
	MergedNodeResultArtifactID        = "artifact_id"
	MergedNodeResultMetadata          = "metadata"
	MergedNodeResultContentPath       = "content_path"
	MergedNodeResultArtifactExecState = "artifact_exec_state"
)

// An MergedNodeResult maps to the merged_node_result table.
type MergedNodeResult struct {
	ID                uuid.UUID                         `db:"id" json:"id"`
	OperatorExecState shared.NullExecutionState         `db:"operator_exec_state" json:"operator_exec_state"`
	ArtifactID        uuid.UUID                         `db:"artifact_id" json:"artifact_id"`
	Metadata          shared.NullArtifactResultMetadata `db:"metadata" json:"metadata"`
	ContentPath       string                            `db:"content_path" json:"content_path"`
	ArtifactExecState shared.NullExecutionState         `db:"artifact_exec_state" json:"artifact_exec_state"`
}

// MergedNodeResultCols returns a comma-separated string of all MergedNodeResult columns.
func MergedNodeResultCols() string {
	return strings.Join(allMergedNodeResultCols(), ",")
}

// MergedNodeResultColsWithPrefix returns a comma-separated string of all
// MergedNodeResult columns prefixed by the table name.
func MergedNodeResultColsWithPrefix() string {
	cols := allMergedNodeResultCols()
	for i, col := range cols {
		cols[i] = fmt.Sprintf("%s.%s", MergedNodeResultTable, col)
	}

	return strings.Join(cols, ",")
}

func allMergedNodeResultCols() []string {
	return []string{
		MergedNodeResultID,
		MergedNodeResultOperatorExecState,
		MergedNodeResultArtifactID,
		MergedNodeResultMetadata,
		MergedNodeResultContentPath,
		MergedNodeResultArtifactExecState,
	}
}
