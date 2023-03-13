package sqlite

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type dagEdgeRepo struct {
	dagEdgeReader
	dagEdgeWriter
}

type dagEdgeReader struct{}

type dagEdgeWriter struct{}

func NewDAGEdgeRepo() repos.DAGEdge {
	return &dagEdgeRepo{
		dagEdgeReader: dagEdgeReader{},
		dagEdgeWriter: dagEdgeWriter{},
	}
}

func (*dagEdgeReader) GetArtifactToOperatorByDAG(
	ctx context.Context,
	dagID uuid.UUID,
	DB database.Database,
) ([]models.DAGEdge, error) {
	query := fmt.Sprintf(
		`SELECT %s 
		FROM workflow_dag_edge 
		WHERE 
			workflow_dag_id = $1 
			AND type = '%s' ORDER BY idx;`,
		models.DAGEdgeCols(),
		shared.ArtifactToOperatorDAGEdge,
	)
	args := []interface{}{dagID}

	return getDAGEdges(ctx, DB, query, args...)
}

func (*dagEdgeReader) GetByDAGBatch(
	ctx context.Context,
	dagIDs []uuid.UUID,
	DB database.Database,
) ([]models.DAGEdge, error) {
	if len(dagIDs) == 0 {
		return nil, errors.New("Provided empty dagIDs list.")
	}

	query := fmt.Sprintf(
		`SELECT %s 
		FROM workflow_dag_edge 
		WHERE workflow_dag_id IN (%s);`,
		models.DAGEdgeCols(),
		stmt_preparers.GenerateArgsList(len(dagIDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(dagIDs)

	return getDAGEdges(ctx, DB, query, args...)
}

func (*dagEdgeReader) GetOperatorToArtifactByDAG(
	ctx context.Context,
	dagID uuid.UUID,
	DB database.Database,
) ([]models.DAGEdge, error) {
	query := fmt.Sprintf(
		`SELECT %s 
		FROM workflow_dag_edge 
		WHERE 
			workflow_dag_id = $1 
			AND type = '%s' ORDER BY idx;`,
		models.DAGEdgeCols(),
		shared.OperatorToArtifactDAGEdge,
	)
	args := []interface{}{dagID}

	return getDAGEdges(ctx, DB, query, args...)
}

func (*dagEdgeWriter) Create(
	ctx context.Context,
	dagID uuid.UUID,
	edgeType shared.DAGEdgeType,
	fromID uuid.UUID,
	toID uuid.UUID,
	idx int16,
	DB database.Database,
) (*models.DAGEdge, error) {
	cols := []string{
		models.DAGEdgeDagID,
		models.DAGEdgeType,
		models.DAGEdgeFromID,
		models.DAGEdgeToID,
		models.DAGEdgeIdx,
	}
	query := DB.PrepareInsertWithReturnAllStmt(models.DAGEdgeTable, cols, models.DAGEdgeCols())

	args := []interface{}{
		dagID,
		edgeType,
		fromID,
		toID,
		idx,
	}

	return getDAGEdge(ctx, DB, query, args...)
}

func (*dagEdgeWriter) DeleteByDAGBatch(ctx context.Context, dagIDs []uuid.UUID, DB database.Database) error {
	query := fmt.Sprintf(
		`DELETE FROM workflow_dag_edge
		WHERE workflow_dag_id IN (%s);`,
		stmt_preparers.GenerateArgsList(len(dagIDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(dagIDs)

	return DB.Execute(ctx, query, args...)
}

func getDAGEdges(ctx context.Context, DB database.Database, query string, args ...interface{}) ([]models.DAGEdge, error) {
	var edges []models.DAGEdge
	err := DB.Query(ctx, &edges, query, args...)
	return edges, err
}

func getDAGEdge(ctx context.Context, DB database.Database, query string, args ...interface{}) (*models.DAGEdge, error) {
	edges, err := getDAGEdges(ctx, DB, query, args...)
	if err != nil {
		return nil, err
	}

	if len(edges) == 0 {
		return nil, database.ErrNoRows()
	}

	if len(edges) != 1 {
		return nil, errors.Newf("Expected 1 DAGEdge but got %v", len(edges))
	}

	return &edges[0], nil
}
