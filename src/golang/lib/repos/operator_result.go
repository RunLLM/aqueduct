package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

// OperatorResult defines all of the database operations that can be performed for a OperatorResult.
type OperatorResult interface {
	operatorResultReader
	operatorResultWriter
}

type operatorResultReader interface {
	// Get returns the OperatorResult with ID.
	Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.OperatorResult, error)

	// GetBatch returns the OperatorResults with IDs.
	GetBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]models.OperatorResult, error)

	// GetByDAGResultAndOperator returns the OperatorResult given the operatorID and workflowDAGResultIDs.
	GetByDAGResultAndOperator(ctx context.Context, workflowDAGResultIDs, operatorID uuid.UUID, DB database.Database) (*models.OperatorResult, error)

	// GetByDAGResults returns the OperatorResult given workflowDAGResultIDs.
	GetByDAGResults(ctx context.Context, workflowDAGResultIDs []uuid.UUID, DB database.Database) ([]models.OperatorResult, error)
}

type operatorResultWriter interface {
	// Create inserts a new OperatorResult with the specified fields.
	Create(
		ctx context.Context,
		dagResultID uuid.UUID,
		operatorID uuid.UUID,
		execState *shared.ExecutionState,
		DB database.Database,
	) (*models.OperatorResult, error)

	// Delete deletes the OperatorResult with ID.
	Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error

	// DeleteBatch deletes the OperatorResults with IDs.
	DeleteBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) error

	// Update applies changes to the OperatorResult with ID. It returns the updated OperatorResult.
	Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, DB database.Database) (*models.OperatorResult, error)
}
