package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator"
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

	// GetNode returns the OperatorNode view given the operator ID.
	GetNode(ctx context.Context, ID uuid.UUID, DB database.Database) (*views.OperatorNode, error)

	// GetNodeBatch returns the OperatorNodes the operator IDs.
	GetNodeBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]views.OperatorNode, error)

	// GetBatch returns the Operators with IDs.
	GetBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]models.Operator, error)

	// GetByDAG returns the Operators in the specified DAG.
	GetByDAG(ctx context.Context, dagID uuid.UUID, DB database.Database) ([]models.Operator, error)

	// GetNodesByDAG returns the OperatorNodes in the specified DAG.
	// OperatorNodes includes inputs / outputs IDs, which are not available in
	// Operators.
	GetNodesByDAG(ctx context.Context, dagID uuid.UUID, DB database.Database) ([]views.OperatorNode, error)

	// GetDistinctLoadOPsByWorkflow returns the distinct Load Operators in a workflow.
	// Load Operators are distinct if they have a unique combination of
	// the resource they are saving an Artifact to, the name of the object (i.e. table)
	// the Artifact is being saved to, and the update mode used to save the Artifact.
	GetDistinctLoadOPsByWorkflow(ctx context.Context, workflowID uuid.UUID, DB database.Database) ([]views.LoadOperator, error)

	// GetExtractAndLoadOPsByResource returns all Extract and Load Operators
	// using the Resource specified.
	GetExtractAndLoadOPsByResource(
		ctx context.Context,
		resourceID uuid.UUID,
		DB database.Database,
	) ([]models.Operator, error)

	// GetEngineTypesByDagID retrieves all engine types keyed by DAG ID,
	// based on the given list of DAG IDs.
	GetEngineTypesMapByDagIDs(
		ctx context.Context,
		DagIDs []uuid.UUID,
		DB database.Database,
	) (map[uuid.UUID][]shared.EngineType, error)

	// GetLoadOPsByWorkflowAndResource returns the Operators in a Workflow related to an Resource.
	GetLoadOPsByWorkflowAndResource(
		ctx context.Context,
		workflowID uuid.UUID,
		resourceID uuid.UUID,
		objectName string,
		DB database.Database,
	) ([]models.Operator, error)

	// GetLoadOPsByResource returns the Operators related to an resource.
	GetLoadOPsByResource(
		ctx context.Context,
		resourceID uuid.UUID,
		objectName string,
		DB database.Database,
	) ([]models.Operator, error)

	// GetLoadOPSpecsByOrg returns a LoadOperatorSpec for each Load Operator in any of the specified
	// organization's Workflows.
	GetLoadOPSpecsByOrg(ctx context.Context, orgID string, DB database.Database) ([]views.LoadOperatorSpec, error)

	// GetRelationBatch returns an OperatorRelation for each Operator in IDs.
	GetRelationBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]views.OperatorRelation, error)

	// GetByEngineType returns all Operators that uses the given engine type.
	GetByEngineType(ctx context.Context, engineType shared.EngineType, DB database.Database) ([]models.Operator, error)

	// GetUnusedCondaEnvNames returns all inactive conda env IDs captured in `engine_config`,
	// if the engine type is AqueductConda.
	GetUnusedCondaEnvNames(ctx context.Context, DB database.Database) ([]string, error)

	// GetByEngineResourceID returns all operators executing on the given engine ID.
	// This includes all operators with engine_config field to be this ID,
	// or those who inherit workflow's engine_config that uses this ID.
	// This does not work with the Aqueduct Engine resource. For that, use `GetForAqueductEngine`.
	GetByEngineResourceID(ctx context.Context, resourceID uuid.UUID, DB database.Database) ([]models.Operator, error)

	// GetForAqueductEngine returns all operators executed on the native Aqueduct Engine.
	GetForAqueductEngine(ctx context.Context, DB database.Database) ([]models.Operator, error)

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
