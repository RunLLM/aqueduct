package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/google/uuid"
)

// Operator defines all of the database operations that can be performed for a operator.
type Operator interface {
	operatorReader
	operatorWriter
}

type operatorReader interface {
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
	Exists(ctx context.Context, id uuid.UUID, db database.Database) (bool, error)
	GetOperator(ctx context.Context, id uuid.UUID, db database.Database) (*DBOperator, error)
	GetOperators(ctx context.Context, ids []uuid.UUID, db database.Database) ([]DBOperator, error)
	GetOperatorsByWorkflowDagId(
		ctx context.Context,
		workflowDagId uuid.UUID,
		db database.Database,
	) ([]DBOperator, error)
	GetDistinctLoadOperatorsByWorkflowId(
		ctx context.Context,
		workflowId uuid.UUID,
		db database.Database,
	) ([]GetDistinctLoadOperatorsByWorkflowIdResponse, error)
	GetLoadOperatorsForWorkflowAndIntegration(
		ctx context.Context,
		workflowId uuid.UUID,
		integrationId uuid.UUID,
		objectName string,
		db database.Database,
	) ([]DBOperator, error)
	ValidateOperatorOwnership(
		ctx context.Context,
		organizationId string,
		operatorId uuid.UUID,
		db database.Database,
	) (bool, error)
	GetOperatorsByIntegrationId(
		ctx context.Context,
		integrationId uuid.UUID,
		db database.Database,
	) ([]DBOperator, error)
}

type operatorWriter interface {
	// Create inserts a new Operator with the specified fields.
	Create(
		ctx context.Context,
		name string,
		description string,
		spec *Spec,
		db database.Database,
	) (*models.Operator, error)

	// Delete deletes the Operator with id.
	Delete(ctx context.Context, id uuid.UUID, db database.Database) error

	// DeleteMultiple deletes the Operators with ids.
	DeleteMultiple(ctx context.Context, ids []uuid.UUID, db database.Database) error

	// Update applies changes to the Operator with id. It returns the updated Operator.
	Update(ctx context.Context, id uuid.UUID, changes map[string]interface{}, db database.Database) (*models.Operator, error)
}
