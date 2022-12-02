package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

// DAGEdge defines all of the database operations that can be performed for a DAGEdge.
type DAGEdge interface {
	dagEdgeReader
	dagEdgeWriter
}

type dagEdgeReader interface {
	// GetArtifactToOperatorByDAG returns all DAGEdges from an Artifact to an Operator for the DAG specified.
	// The DAGEdges are ordered by idx.
	GetArtifactToOperatorByDAG(ctx context.Context, dagID uuid.UUID, DB database.Database) ([]models.DAGEdge, error)

	// GetByDAGBatch returns all DAGEdges for the DAGs specified.
	GetByDAGBatch(ctx context.Context, dagIDs []uuid.UUID, DB database.Database) ([]models.DAGEdge, error)

	// GetOperatorToArtifactByDAG returns all DAGEdges from an Operator to an Artifact for the DAG specified.
	// The DAGEdges are ordered by idx.
	GetOperatorToArtifactByDAG(ctx context.Context, dagID uuid.UUID, DB database.Database) ([]models.DAGEdge, error)
}

type dagEdgeWriter interface {
	// Create inserts a new DAGEdge with the fields specified.
	Create(
		ctx context.Context,
		dagID uuid.UUID,
		edgeType shared.DAGEdgeType,
		fromID uuid.UUID,
		toID uuid.UUID,
		idx int16,
		DB database.Database,
	) (*models.DAGEdge, error)

	// DeleteByDAGBatch deletes all DAGEdges of the DAGs with ID in dagIDs.
	DeleteByDAGBatch(ctx context.Context, dagIDs []uuid.UUID, DB database.Database) error
}
