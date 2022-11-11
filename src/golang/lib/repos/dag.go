package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

type DAG interface {
	dagReader
	dagWriter
}

type dagReader interface {
	// Get returns the DAG with ID.
	// It returns a database.ErrNoRows if no rows are found.
	Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.DAG, error)

	// GetBatch returns the DAGs with ID in IDs.
	GetBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]models.DAG, error)

	// GetByArtifactResultBatch returns the DAG used to create the ArtifactResult for each
	// ArtifactResult specified. It returns a map of the ArtifactResult.ID to the DAG.
	GetByArtifactResultBatch(ctx context.Context, artifactResultIDs []uuid.UUID, DB database.Database) (map[uuid.UUID]models.DAG, error)

	// GetByDAGResults returns the DAG used to create the DAGResult specified.
	GetByDAGResult(ctx context.Context, dagResultID uuid.UUID, DB database.Database) (*models.DAG, error)

	// GetByOperator returns the DAG that the specified Operator belongs to.
	GetByOperator(ctx context.Context, operatorID uuid.UUID, DB database.Database)

	// GetByWorkflow returns all DAGs for the Workflow specified.
	GetByWorkflow(ctx context.Context, workflowID uuid.UUID, DB database.Database) ([]models.DAG, error)

	// GetLatestByWorkflow returns the latest DAG for the Workflow specified.
	GetLatestByWorkflow(ctx context.Context, workflowID uuid.UUID, DB database.Database) (*models.DAG, error)
}

type dagWriter interface {
	// Create inserts a new DAG with the specified fields.
	Create(
		ctx context.Context,
		workflowID uuid.UUID,
		storageConfig *shared.StorageConfig,
		engineConfig *shared.EngineConfig,
		DB database.Database,
	) (*models.DAG, error)

	// Delete deletes the DAG with ID.
	Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error

	// DeleteBatch deletes all DAGs with ID in IDs.
	DeleteBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) error

	// Update applies changes to the DAG with ID. It returns the updated DAG.
	Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, DB database.Database)
}
