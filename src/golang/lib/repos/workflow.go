package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	mdl_shared "github.com/aqueducthq/aqueduct/lib/models/shared"
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
	// It returns a database.ErrNoRows if no rows are found.
	Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.Workflow, error)

	// GetByDAG returns the Workflow associated with the specified DAG.
	// It returns a database.ErrNoRows if no rows are found.
	GetByDAG(ctx context.Context, dagID uuid.UUID, DB database.Database) (*models.Workflow, error)

	// GetByOwnerAndName returns the workflow created by ownerID named name.
	// It returns a database.ErrNoRows if no rows are found.
	GetByOwnerAndName(ctx context.Context, ownerID uuid.UUID, name string, DB database.Database) (*models.Workflow, error)

	// GetByScheduleTrigger returns all Workflows where Schedule.Trigger is equal to trigger.
	GetByScheduleTrigger(ctx context.Context, trigger workflow.UpdateTrigger, DB database.Database) ([]models.Workflow, error)

	// GetTargets returns the ID of each Workflow where Schedule.SourceID
	// is equal to ID.
	GetTargets(ctx context.Context, ID uuid.UUID, DB database.Database) ([]uuid.UUID, error)

	// GetLastRunByEngine returns a WorkflowLastRun for each Workflow where the latest
	// DAGResult created is associated with a DAG running on engine.
	GetLastRunByEngine(ctx context.Context, engine shared.EngineType, DB database.Database) ([]views.WorkflowLastRun, error)

	// GetLatestStatusesByOrg returns the LatestWorkflowStatus for each workflow owned by orgID.
	GetLatestStatusesByOrg(ctx context.Context, orgID string, DB database.Database) ([]views.LatestWorkflowStatus, error)

	// List returns all Workflows.
	List(ctx context.Context, DB database.Database) ([]models.Workflow, error)

	// ValidateOrg returns whether the Workflow was created by a user in orgID.
	ValidateOrg(ctx context.Context, ID uuid.UUID, orgID string, DB database.Database) (bool, error)
}

type workflowWriter interface {
	// Create inserts a new Workflow with the specified fields.
	Create(
		ctx context.Context,
		userID uuid.UUID,
		name string,
		description string,
		schedule *workflow.Schedule,
		retentionPolicy *workflow.RetentionPolicy,
		notificationSettings mdl_shared.NotificationSettings,
		DB database.Database,
	) (*models.Workflow, error)

	// Delete deletes the Workflow with ID.
	Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error

	// Update applies changes to the Workflow with ID. It returns the updated Workflow.
	Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, DB database.Database) (*models.Workflow, error)
}
