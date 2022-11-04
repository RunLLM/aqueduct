package models

import (
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

// An Artifact maps to the artifact table.
type Artifact struct {
	ID          uuid.UUID   `db:"id" json:"id"`
	Name        string      `db:"name" json:"name"`
	Description string      `db:"description" json:"description"`
	Type        shared.ArtifactType `db:"type" json:"type"`
}
