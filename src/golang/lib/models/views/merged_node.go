package views

import (
	"fmt"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator"
	"github.com/google/uuid"
)

const (
	OperatorWithArtifactNodeView          = "merged_node"
	OperatorWithArtifactNodeID            = "id"
	OperatorWithArtifactNodeDagID         = "dag_id"
	OperatorWithArtifactNodeArtifactID    = "artifact_id"
	OperatorWithArtifactNodeName          = "name"
	OperatorWithArtifactNodeDescription   = "description"
	OperatorWithArtifactNodeSpec          = "spec"
	OperatorWithArtifactNodeType          = "type"
	OperatorWithArtifactNodeInputs        = "inputs"
	OperatorWithArtifactNodeOutputs       = "outputs"
	OperatorWithArtifactNodeShouldPersist = "should_persist"
)

type OperatorWithArtifactNode struct {
	ID            uuid.UUID           `db:"id" json:"id"`
	DagID         uuid.UUID           `db:"dag_id" json:"dag_id"`
	ArtifactID    uuid.UUID           `db:"artifact_id" json:"artifact_id"`
	Name          string              `db:"name" json:"name"`
	Description   string              `db:"description" json:"description"`
	Spec          operator.Spec       `db:"spec" json:"spec"`
	Type          shared.ArtifactType `db:"type" json:"type"`
	ShouldPersist bool                `db:"should_persist" json:"should_persist"`

	Inputs  shared.NullableIndexedList[uuid.UUID] `db:"inputs" json:"inputs"`
	Outputs shared.NullableIndexedList[uuid.UUID] `db:"outputs" json:"outputs"`
}

// OperatorWithArtifactNodeCols returns a comma-separated string of all merged node columns.
func OperatorWithArtifactNodeCols() string {
	return strings.Join(allOperatorWithArtifactNodeCols(), ",")
}

// OperatorWithArtifactNodeColsWithPrefix returns a comma-separated string of all
// merged node columns prefixed by the view name.
func OperatorWithArtifactNodeColsWithPrefix() string {
	cols := allOperatorWithArtifactNodeCols()
	for i, col := range cols {
		cols[i] = fmt.Sprintf("%s.%s", OperatorWithArtifactNodeView, col)
	}

	return strings.Join(cols, ",")
}

func allOperatorWithArtifactNodeCols() []string {
	return []string{
		OperatorWithArtifactNodeID,
		OperatorWithArtifactNodeDagID,
		OperatorWithArtifactNodeArtifactID,
		OperatorWithArtifactNodeName,
		OperatorWithArtifactNodeDescription,
		OperatorWithArtifactNodeSpec,
		OperatorWithArtifactNodeType,
		OperatorWithArtifactNodeInputs,
		OperatorWithArtifactNodeOutputs,
		OperatorWithArtifactNodeShouldPersist,
	}
}
