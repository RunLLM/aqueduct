package views

import (
	"fmt"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

const (
	OperatorWithArtifactResultNodeTable = "operator_with_artifact_node_result"

	// OperatorWithArtifactResultNode table column names
	OperatorWithArtifactResultNodeID                = "id"  // operator result ID
	OperatorWithArtifactResultNodeArtifactResultID        = "artifact_result_id"
	OperatorWithArtifactResultNodeOperatorID        = "operator_id"
	OperatorWithArtifactResultNodeArtifactID        = "artifact_id"
	OperatorWithArtifactResultNodeOperatorResultExecState = "operator_result_exec_state"
	OperatorWithArtifactResultNodeMetadata          = "metadata"
	OperatorWithArtifactResultNodeContentPath       = "content_path"
	OperatorWithArtifactResultNodeArtifactResultExecState = "artifact_result_exec_state"
)

// An OperatorWithArtifactResultNode maps to the merged_node_result table.
type OperatorWithArtifactResultNode struct {
	ID                uuid.UUID                         `db:"id" json:"id"`
	OperatorID        uuid.UUID                         `db:"operator_id" json:"operator_id"`
	OperatorResultExecState shared.NullExecutionState         `db:"operator_result_exec_state" json:"operator_result_exec_state"`
	ArtifactID        uuid.UUID                         `db:"artifact_id" json:"artifact_id"`
	ArtifactResultID        uuid.UUID                         `db:"artifact_result_id" json:"artifact_result_id"`
	Metadata          shared.NullArtifactResultMetadata `db:"metadata" json:"metadata"`
	ContentPath       string                            `db:"content_path" json:"content_path"`
	ArtifactResultExecState shared.NullExecutionState         `db:"artifact_result_exec_state" json:"artifact_result_exec_state"`
}

// OperatorWithArtifactResultNodeCols returns a comma-separated string of all OperatorWithArtifactResultNode columns.
func OperatorWithArtifactResultNodeCols() string {
	return strings.Join(allOperatorWithArtifactResultNodeCols(), ",")
}

// OperatorWithArtifactResultNodeColsWithPrefix returns a comma-separated string of all
// OperatorWithArtifactResultNode columns prefixed by the table name.
func OperatorWithArtifactResultNodeColsWithPrefix() string {
	cols := allOperatorWithArtifactResultNodeCols()
	for i, col := range cols {
		cols[i] = fmt.Sprintf("%s.%s", OperatorWithArtifactResultNodeTable, col)
	}

	return strings.Join(cols, ",")
}

func allOperatorWithArtifactResultNodeCols() []string {
	return []string{
		OperatorWithArtifactResultNodeID,
		OperatorWithArtifactResultNodeOperatorResultExecState,
		OperatorWithArtifactResultNodeOperatorID,
		OperatorWithArtifactResultNodeArtifactID,
		OperatorWithArtifactResultNodeArtifactResultID,
		OperatorWithArtifactResultNodeMetadata,
		OperatorWithArtifactResultNodeContentPath,
		OperatorWithArtifactResultNodeArtifactResultExecState,
	}
}
