package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/google/uuid"
)

// Operator defines all of the database operations that can be performed for a operator.
type Operator interface {
	operatorReader
	operatorWriter
}

type operatorReader interface {
	// Exists returns whether a Operator with ID exists.
	Exists(ctx context.Context, ID uuid.UUID, DB database.Database) (bool, error)

	// Get returns the Operator with ID.
	Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.Operator, error)

	// GetBatch returns the Operators with IDs.
	GetBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]models.Operator, error)

	// GetByDAG returns the Operators in the specified DAG.
	GetByDAG(ctx context.Context, dagID uuid.UUID, DB database.Database) ([]models.Operator, error)

	// GetDistinctLoadOPsByWorkflow returns the distinct Load Operators in a workflow.
	// Load Operators are distinct if they have a unique combination of
	// the integration they are saving an Artifact to, the name of the object (i.e. table)
	// the Artifact is being saved to, and the update mode used to save the Artifact.
	GetDistinctLoadOPsByWorkflow(ctx context.Context, workflowID uuid.UUID, DB database.Database) ([]views.LoadOperator, error)

	// GetLoadOPsByWorkflowAndIntegration returns the Operators in a Workflow related to an Integration.
	GetLoadOPsByWorkflowAndIntegration(
		ctx context.Context,
		workflowID uuid.UUID,
		integrationID uuid.UUID,
		objectName string,
		DB database.Database,
	) ([]models.Operator, error)

	// GetLoadOPsByIntegration returns the Operators related to an integration.
	GetLoadOPsByIntegration(
		ctx context.Context,
		integrationID uuid.UUID,
		objectName string,
		DB database.Database,
	) ([]models.Operator, error)

	// GetLoadOPSpecsByOrg returns a LoadOperatorSpec for each Load Operator in any of the specified
	// organization's Workflows.
	GetLoadOPSpecsByOrg(ctx context.Context, orgID string, DB database.Database) ([]views.LoadOperatorSpec, error)

	// GetRelationBatch returns an OperatorRelation for each Operator in IDs.
	GetRelationBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]views.OperatorRelation, error)

	// ValidateOrg returns whether the Operator was created by the specified organization.
	ValidateOrg(ctx context.Context, ID uuid.UUID, orgID string, DB database.Database) (bool, error)
}

type operatorWriter interface {
	// Create inserts a new Operator with the specified fields.
	Create(
		ctx context.Context,
		name string,
		description string,
		spec *operator.Spec,
		executionEnvironmentID *uuid.UUID,
		DB database.Database,
	) (*models.Operator, error)

	// Delete deletes the Operator with ID.
	Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error

	// DeleteBatch deletes the Operators with IDs.
	DeleteBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) error

	// Update applies changes to the Operator with ID. It returns the updated Operator.
	Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, DB database.Database) (*models.Operator, error)
}
