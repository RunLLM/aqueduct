package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/google/uuid"
)

// DAGResult defines all of the database operations that can be performed for a DAGResult.
type DAGResult interface {
	dagResultReader
	dagResultWriter
}

type dagResultReader interface {
	// Get returns the DAGResult with ID.
	// It returns a database.ErrNoRows if no rows are found.
	Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.DAGResult, error)

	// GetBatch returns the DAGResults with ID in IDs.
	GetBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]models.DAGResult, error)

	// GetByWorkflow returns the DAGResults of all DAGs associated with the Workflow with workflowID.
	GetByWorkflow(ctx context.Context, workflowID uuid.UUID, DB database.Database) ([]models.DAGResult, error)

	// GetKOffsetByWorkflow returns the DAGResults of all DAGs associated with the Workflow with workflowID
	// except for the last k DAGResults ordered by DAGResult.CreatedAt.
	GetKOffsetByWorkflow(ctx context.Context, workflowID uuid.UUID, k int, DB database.Database) ([]models.DAGResult, error)

	// GetWorkflowMetadataBatch returns a map of each DAGResult's ID in IDs to its DAGResultWorkflowMetadata.
	GetWorkflowMetadataBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) (map[uuid.UUID]views.DAGResultWorkflowMetadata, error)
}

type dagResultWriter interface {
	// Creates inserts a new DAGResult with the specified fields.
	// It returns an ErrInvalidPendingTimestamp if execState.Timestamps.PendingAt is
	// not set, since that value is used for DAGResult.CreatedAt.
	Create(
		ctx context.Context,
		dagID uuid.UUID,
		execState *shared.ExecutionState,
		DB database.Database,
	) (*models.DAGResult, error)

	// Delete deletes the DAGResult with ID.
	Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error

	// DeleteBatch deletes all DAGResults with ID in IDs.
	DeleteBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) error

	// Update applies changes to the the DAGResult with ID. It returns the updated DAGResult.
	Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, DB database.Database) (*models.DAGResult, error)
}
