package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

// Artifact defines all of the database operations that can be performed for an Artifact.
type Artifact interface {
	artifactReader
	artifactWriter
}

type artifactReader interface {
	// Exists returns whether an Artifact with ID exists.
	Exists(ctx context.Context, ID uuid.UUID, db database.Database) (bool, error)

	// Get returns the Artifact with ID.
	Get(ctx context.Context, ID uuid.UUID, db database.Database) (*models.Artifact, error)

	// Get returns the Artifacts with IDs.
	GetMultiple(ctx context.Context, IDs []uuid.UUID, db database.Database) ([]models.Artifact, error)

	// GetByWorkflowDagId returns the workflows created by workflow dag with ID workflowDagId.
	GetByWorkflowDagId(ctx context.Context, workflowDagID uuid.UUID, db database.Database) ([]models.Artifact, error)

	// ValidateOrg returns whether the Workflow was created by a user in orgID.
	ValidateOrg(ctx context.Context, ID uuid.UUID, orgID uuid.UUID, db database.Database) (bool, error)
}

type artifactWriter interface {
	// Create inserts a new Artifact with the specified fields.
	Create(
		ctx context.Context,
		name string,
		description string,
		artifactType shared.Type,
		db database.Database,
	) (*models.Artifact, error)

	// Update applies changes to the Artifact with ID. It returns the updated Artifact.
	Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, db database.Database) (*models.Artifact, error)

	// Delete deletes the Artifact with ID.
	Delete(ctx context.Context, ID uuid.UUID, db database.Database) error

	// DeleteMultiple deletes the Artifacts with IDs.
	DeleteMultiple(ctx context.Context, IDs []uuid.UUID, db database.Database) error
}
