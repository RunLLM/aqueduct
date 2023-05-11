package views

import (
	"fmt"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

const (
	MergedNodeView    = "merged_node"
	MergedNodeID = "id"
	MergedNodeDagID = "dag_id"
	MergedNodeArtifactID = "artifact_id"
	MergedNodeName = "name"
	MergedNodeDescription = "description"
	MergedNodeSpec = "spec"
	MergedNodeType = "type"
	MergedNodeInputs = "inputs"
	MergedNodeOutputs = "outputs"
)

type MergedNode struct {
	ID          uuid.UUID           `db:"id" json:"id"`
	DagID       uuid.UUID           `db:"dag_id" json:"dag_id"`
	ArtifactID          uuid.UUID           `db:"artifact_id" json:"artifact_id"`
	Name        string              `db:"name" json:"name"`
	Description string              `db:"description" json:"description"`
	Spec                   operator.Spec  `db:"spec" json:"spec"`
	Type        shared.ArtifactType `db:"type" json:"type"`

	Inputs   shared.NullableIndexedList[uuid.UUID]                             `db:"inputs" json:"inputs"`
	Outputs shared.NullableIndexedList[uuid.UUID] `db:"outputs" json:"outputs"`
}

// MergedNodeCols returns a comma-separated string of all merged node columns.
func MergedNodeCols() string {
	return strings.Join(allMergedNodeCols(), ",")
}

// MergedNodeColsWithPrefix returns a comma-separated string of all
// merged node columns prefixed by the view name.
func MergedNodeColsWithPrefix() string {
	cols := allMergedNodeCols()
	for i, col := range cols {
		cols[i] = fmt.Sprintf("%s.%s", MergedNodeView, col)
	}

	return strings.Join(cols, ",")
}

func allMergedNodeCols() []string {
	mergedNodeCols = append(
		MergedNodeID,
		MergedNodeDagID,
		MergedNodeArtifactID,
		MergedNodeName,
		MergedNodeDescription,
		MergedNodeSpec,
		MergedNodeType,
		MergedNodeInputs,
		MergedNodeOutputs,
	)

	return mergedNodeCols
}
