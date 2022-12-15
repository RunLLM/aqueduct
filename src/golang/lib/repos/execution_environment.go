package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

// ExecutionEnvironment defines all of the database operations that can be performed for a ExecutionEnvironment.
type ExecutionEnvironment interface {
	executionEnvironmentReader
	executionEnvironmentWriter
}

type executionEnvironmentReader interface {
	// Get returns the ExecutionEnvironment with ID.
	Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.ExecutionEnvironment, error)

	// GetBatch returns the ExecutionEnvironments with IDs.
	GetBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]models.ExecutionEnvironment, error)

	// GetActiveByHash returns the environments which aren't garbage-collected by the hash.
	GetActiveByHash(ctx context.Context, hash uuid.UUID, DB database.Database) (*models.ExecutionEnvironment, error)

	// GetActiveByOperatorBatch returns the environments which aren't garbage-collected that is being used by the operators specified by the opIDs.
	// This is done in key-value format where the key is the operator ID and the value is the environment.
	GetActiveByOperatorBatch(ctx context.Context, opIDs []uuid.UUID, DB database.Database) (map[uuid.UUID]models.ExecutionEnvironment, error)

	// GetUnused returns the environments which aren't used by operators in the latest DAG of a workflow.
	GetUnused(ctx context.Context, DB database.Database) ([]models.ExecutionEnvironment, error)
}

type executionEnvironmentWriter interface {
	// Create inserts a new ExecutionEnvironment with the specified fields.
	Create(
		ctx context.Context,
		spec *shared.ExecutionEnvironmentSpec,
		hash uuid.UUID,
		DB database.Database,
	) (*models.ExecutionEnvironment, error)

	// Delete deletes the ExecutionEnvironment with ID.
	Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error

	// DeleteBatch deletes the ExecutionEnvironment with ID.
	DeleteBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) error

	// Update applies changes to the ExecutionEnvironment with ID. It returns the updated ExecutionEnvironment.
	Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, DB database.Database) (*models.ExecutionEnvironment, error)
}
