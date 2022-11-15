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
	// Exists returns whether a Workflow with ID exists.
	Exists(ctx context.Context, ID uuid.UUID, DB database.Database) (bool, error)

	// Get returns the Workflow with ID.
<<<<<<< HEAD
	Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.Workflow, error)

	// GetByOwnerAndName returns the workflow created by ownerID named name.
	GetByOwnerAndName(ctx context.Context, ownerID uuid.UUID, name string, DB database.Database) (*models.Workflow, error)

	// GetLatestStatusesByOrg returns the LatestWorkflowStatus for each workflow owned by orgID.
	GetLatestStatusesByOrg(ctx context.Context, orgID uuid.UUID, DB database.Database) ([]views.LatestWorkflowStatus, error)
=======
	// It returns a database.ErrNoRows if no rows are found.
	Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.Workflow, error)

	// GetByOwnerAndName returns the workflow created by ownerID named name.
	// It returns a database.ErrNoRows if no rows are found.
	GetByOwnerAndName(ctx context.Context, ownerID uuid.UUID, name string, DB database.Database) (*models.Workflow, error)

	// GetLatestStatusesByOrg returns the LatestWorkflowStatus for each workflow owned by orgID.
	GetLatestStatusesByOrg(ctx context.Context, orgID string, DB database.Database) ([]views.LatestWorkflowStatus, error)
>>>>>>> 5b030decaf256156468b4ac0ae184775aedee738

	// List returns all Workflows.
	List(ctx context.Context, DB database.Database) ([]models.Workflow, error)

	// ValidateOrg returns whether the Workflow was created by a user in orgID.
<<<<<<< HEAD
	ValidateOrg(ctx context.Context, ID uuid.UUID, orgID uuid.UUID, DB database.Database) (bool, error)
=======
	ValidateOrg(ctx context.Context, ID uuid.UUID, orgID string, DB database.Database) (bool, error)
>>>>>>> 5b030decaf256156468b4ac0ae184775aedee738
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
		DB database.Database,
	) (*models.Workflow, error)

	// Delete deletes the Workflow with ID.
	Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error

	// Update applies changes to the Workflow with ID. It returns the updated Workflow.
	Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, DB database.Database) (*models.Workflow, error)
}
