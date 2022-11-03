package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/google/uuid"
)

// OperatorResult defines all of the database operations that can be performed for a OperatorResult.
type OperatorResult interface {
	operatorResultReader
	operatorResultWriter
}

type operatorResultReader interface {
	// Get returns the OperatorResult with id.
	Get(ctx context.Context, id uuid.UUID, db database.Database) (*models.OperatorResult, error)

	// GetMultiple returns the OperatorResults with ids.
	GetMultiple(ctx context.Context, ids []uuid.UUID, db database.Database) ([]models.OperatorResult, error)

	// GetByWorkflowDagResultIdAndOperatorId returns the OperatorResult given the operatorId and workflowDagResultId.
	GetByWorkflowDagResultIdAndOperatorId(ctx context.Context, workflowDagResultId, operatorId uuid.UUID, db database.Database) (*models.OperatorResult, error)

	// GetByWorkflowDagResultIds returns the OperatorResult given workflowDagResultIds.
	GetByWorkflowDagResultIds(ctx context.Context, workflowDagResultIds []uuid.UUID, db database.Database) ([]models.OperatorResult, error)
}

type operatorResultWriter interface {
	// Create inserts a new OperatorResult with the specified fields.
	Create(
		ctx context.Context,
		workflowDagResultId uuid.UUID,
		operatorId uuid.UUID,
		execState *shared.ExecutionState,
		db database.Database,
	) (*models.OperatorResult, error)

	// Delete deletes the OperatorResult with id.
	Delete(ctx context.Context, id uuid.UUID, db database.Database) error

	// DeleteMultiple deletes the OperatorResults with ids.
	DeleteMultiple(ctx context.Context, ids []uuid.UUID, db database.Database) error

	// Update applies changes to the OperatorResult with id. It returns the updated OperatorResult.
	Update(ctx context.Context, id uuid.UUID, changes map[string]interface{}, db database.Database) (*models.OperatorResult, error)
}
