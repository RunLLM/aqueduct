package views

import (
	"fmt"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

const (
	ArtifactNodeView    = "artifact_node"
	ArtifactNodeDagID   = "dag_id"
	ArtifactNodeInputs  = "inputs"
	ArtifactNodeOutputs = "outputs"
)

type ArtifactNode struct {
	ID          uuid.UUID           `db:"id" json:"id"`
	DagID       uuid.UUID           `db:"dag_id" json:"dag_id"`
	Name        string              `db:"name" json:"name"`
	Description string              `db:"description" json:"description"`
	Type        shared.ArtifactType `db:"type" json:"type"`

	Inputs  shared.NullableIndexedList[uuid.UUID] `db:"inputs" json:"inputs"`
	Outputs shared.NullableIndexedList[uuid.UUID] `db:"outputs" json:"outputs"`
}

// ArtifactNodeCols returns a comma-separated string of all artifact columns.
func ArtifactNodeCols() string {
	return strings.Join(allArtifactNodeCols(), ",")
}

// ArtifactNodeColsWithPrefix returns a comma-separated string of all
// artifact columns prefixed by the view name.
func ArtifactNodeColsWithPrefix() string {
	cols := allArtifactNodeCols()
	for i, col := range cols {
		cols[i] = fmt.Sprintf("%s.%s", ArtifactNodeView, col)
	}

	return strings.Join(cols, ",")
}

func allArtifactNodeCols() []string {
	artfNodeCols := models.AllArtifactCols()
	artfNodeCols = append(
		artfNodeCols,
		ArtifactNodeDagID,
		ArtifactNodeInputs,
		ArtifactNodeOutputs,
	)

	return artfNodeCols
}
