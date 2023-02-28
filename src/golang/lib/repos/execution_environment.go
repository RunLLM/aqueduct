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

	// GetByHash returns the environments which aren't garbage-collected by the hash.
	GetByHash(ctx context.Context, hash uuid.UUID, DB database.Database) (*models.ExecutionEnvironment, error)

	// GetByOperatorBatch returns the environments which aren't garbage-collected that is being used by the operators specified by the opIDs.
	// This is done in key-value format where the key is the operator ID and the value is the environment.
	GetByOperatorBatch(ctx context.Context, opIDs []uuid.UUID, DB database.Database) (map[uuid.UUID]models.ExecutionEnvironment, error)
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
