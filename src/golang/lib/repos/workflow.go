package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/google/uuid"
)

// Workflow defines all of the database operations that can be performed for a Workflow.
type Workflow interface {
	workflowReader
	workflowWriter
}

type workflowReader interface {
	// Exists returns whether a Workflow with id exists.
	Exists(ctx context.Context, id uuid.UUID, db database.Database) (bool, error)

	// Get returns the Workflow with id.
	Get(ctx context.Context, id uuid.UUID, db database.Database) (*models.Workflow, error)

	// GetByOwnerAndName returns the workflow created by ownerID named name.
	GetByOwnerAndName(ctx context.Context, ownerID uuid.UUID, name string, db database.Database) (*models.Workflow, error)

	// GetLatestStatusesByOrg returns the LatestWorkflowStatus for each workflow owned by orgID.
	GetLatestStatusesByOrg(ctx context.Context, orgID uuid.UUID, db database.Database) ([]views.LatestWorkflowStatus, error)

	// List returns all Workflows.
	List(ctx context.Context, db database.Database) ([]models.Workflow, error)

	// ValidateOrg returns whether the Workflow was created by a user in orgID.
	ValidateOrg(ctx context.Context, id uuid.UUID, orgID uuid.UUID, db database.Database) (bool, error)
}

type workflowWriter interface {
	// Create inserts a new Workflow with the specified fields.
	Create(
		ctx context.Context,
		userID uuid.UUID,
		name string,
		description string,
		schedule *shared.Schedule,
		retentionPolicy *shared.RetentionPolicy,
		db database.Database,
	) (*models.Workflow, error)

	// Delete deletes the Workflow with id.
	Delete(ctx context.Context, id uuid.UUID, db database.Database) error

	// Update applies changes to the Workflow with id. It returns the updated Workflow.
	Update(ctx context.Context, id uuid.UUID, changes map[string]interface{}, db database.Database) (*models.Workflow, error)
}
