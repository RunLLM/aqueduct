package models

import (
	"fmt"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

const (
	ArtifactTable = "artifact"

	// Artifact column names
	ArtifactID          = "id"
	ArtifactName        = "name"
	ArtifactDescription = "description"
	ArtifactType        = "type"
)

// An Artifact maps to the artifact table.
type Artifact struct {
	ID          uuid.UUID           `db:"id" json:"id"`
	Name        string              `db:"name" json:"name"`
	Description string              `db:"description" json:"description"`
	Type        shared.ArtifactType `db:"type" json:"type"`
}

// ArtifactCols returns a comma-separated string of all Artifact columns.
func ArtifactCols() string {
	return strings.Join(allArtifactCols(), ",")
}

func allArtifactCols() []string {
	return []string{
		ArtifactID,
		ArtifactName,
		ArtifactDescription,
		ArtifactType,
	}
}

// ArtifactColsWithPrefix returns a comma-separated string of all
// Artifact columns prefixed by the table name.
func ArtifactColsWithPrefix() string {
	cols := allArtifactCols()
	for i, col := range cols {
		cols[i] = fmt.Sprintf("%s.%s", ArtifactTable, col)
	}

	return strings.Join(cols, ",")
}
