package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/views"
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

	// GetByDAGResultAndOperator returns the OperatorResult for the DAGResult and Operator specified.
	GetByDAGResultAndOperator(ctx context.Context, dagResultID, operatorID uuid.UUID, DB database.Database) (*models.OperatorResult, error)

	// GetByDAGResultBatch returns all OperatorResults for the DAGResults specified.
	GetByDAGResultBatch(ctx context.Context, dagResultIDs []uuid.UUID, DB database.Database) ([]models.OperatorResult, error)

	// GetCheckStatusByArtifactBatch returns an OperatorResultStatus for all OperatorResults
	// associated with a Check Operator where the Operator has incoming DAGEdge
	// from an Artifact in artifactIDs.
	GetCheckStatusByArtifactBatch(
		ctx context.Context,
		artifactIDs []uuid.UUID,
		DB database.Database,
	) ([]views.OperatorResultStatus, error)

	// GetStatusByDAGResultAndArtifactBatch returns an OperatorResultStatus for each
	// OperatorResult belonging to a DAGResult in dagResultIDs where the OperatorResult
	// corresponds to an Operator that has an outgoing DAGEdge to an Artifact in artifactIDs.
	GetStatusByDAGResultAndArtifactBatch(
		ctx context.Context,
		dagResultIDs []uuid.UUID,
		artifactIDs []uuid.UUID,
		DB database.Database,
	) ([]views.OperatorResultStatus, error)
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
