package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

// Operator defines all of the database operations that can be performed for a operator.
type Operator interface {
	operatorReader
	operatorWriter
}

type operatorReader interface {
	// Exists returns whether a Operator with ID exists.
	Exists(ctx context.Context, ID uuid.UUID, db database.Database) (bool, error)

	// Get returns the Operator with ID.
	Get(ctx context.Context, ID uuid.UUID, db database.Database) (*models.Operator, error)

	// GetBatch returns the Operators with IDs.
	GetBatch(ctx context.Context, IDs []uuid.UUID, db database.Database) ([]models.Operator, error)

	// GetByDAG returns the Operators in a workflow DAG.
	GetByDAG(ctx context.Context, workflowDAGID uuid.UUID, db database.Database) ([]models.Operator, error)

	// GetDistinctLoadOperatorsByWorkflow returns the Load Operators in a workflow.
	GetDistinctLoadOperatorsByWorkflow(ctx context.Context, workflowID uuid.UUID, db database.Database) ([]models.LoadOperator, error)

	// GetLoadOperatorsByWorkflowAndIntegration returns the Operators in a workflow related to an integration.
	GetLoadOperatorsByWorkflowAndIntegration(ctx context.Context, workflowID uuid.UUID, integrationID uuid.UUID, objectName string, db database.Database) ([]models.Operator, error)

	// GetLoadOperatorsByIntegration returns the Operators related to an integration.
	GetLoadOperatorsByIntegration(ctx context.Context, integrationID uuid.UUID, objectName string, db database.Database) ([]models.Operator, error)

	// ValidateOrg returns whether the Operator was created by the specified organization.
	ValidateOrg(ctx context.Context, operatorId uuid.UUID, orgID uuid.UUID, db database.Database) (bool, error)
}

type operatorWriter interface {
	// Create inserts a new Operator with the specified fields.
	Create(
		ctx context.Context,
		name string,
		description string,
		spec *shared.Spec,
		db database.Database,
	) (*models.Operator, error)

	// Delete deletes the Operator with ID.
	Delete(ctx context.Context, ID uuid.UUID, db database.Database) error

	// DeleteBatch deletes the Operators with IDs.
	DeleteBatch(ctx context.Context, IDs []uuid.UUID, db database.Database) error

	// Update applies changes to the Operator with ID. It returns the updated Operator.
	Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, db database.Database) (*models.Operator, error)
}
