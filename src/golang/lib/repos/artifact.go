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
	// Exists returns whether an Artifact with id exists.
	Exists(ctx context.Context, id uuid.UUID, db database.Database) (bool, error)

	// Get returns the Artifact with id.
	Get(ctx context.Context, id uuid.UUID, db database.Database) (*models.Artifact, error)

	// Get returns the Artifacts with ids.
	GetMultiple(ctx context.Context, ids []uuid.UUID, db database.Database) ([]models.Artifact, error)

	// GetByWorkflowDagId returns the workflows created by workflow dag with id workflowDagId.
	GetByWorkflowDagId(ctx context.Context, workflowDagId uuid.UUID, db database.Database) ([]models.Artifact, error)

	// ValidateOrg returns whether the Workflow was created by a user in orgID.
	ValidateOrg(ctx context.Context, id uuid.UUID, orgID uuid.UUID, db database.Database) (bool, error)
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

	// Update applies changes to the Artifact with id. It returns the updated Artifact.
	Update(ctx context.Context, id uuid.UUID, changes map[string]interface{}, db database.Database) (*models.Artifact, error)

	// Delete deletes the Artifact with id.
	Delete(ctx context.Context, id uuid.UUID, db database.Database) error

	// DeleteMultiple deletes the Artifacts with ids.
	DeleteMultiple(ctx context.Context, ids []uuid.UUID, db database.Database) error
}
