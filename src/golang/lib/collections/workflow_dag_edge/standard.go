package workflow_dag_edge

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type standardReaderImpl struct{}

type standardWriterImpl struct{}

func (w *standardWriterImpl) CreateWorkflowDagEdge(
	ctx context.Context,
	workflowDagId uuid.UUID,
	edgeType Type,
	fromId uuid.UUID,
	toId uuid.UUID,
	idx int16,
	db database.Database,
) (*WorkflowDagEdge, error) {
	insertColumns := []string{WorkflowDagIdColumn, TypeColumn, FromIdColumn, ToIdColumn, IdxColumn}
	insertWorkflowDagStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	args := []interface{}{workflowDagId, edgeType, fromId, toId, idx}

	var workflowDagEdge WorkflowDagEdge
	err := db.Query(ctx, &workflowDagEdge, insertWorkflowDagStmt, args...)
	return &workflowDagEdge, err
}

func (r *standardReaderImpl) GetOperatorToArtifactEdges(
	ctx context.Context,
	workflowDagId uuid.UUID,
	db database.Database,
) ([]WorkflowDagEdge, error) {
	// We need to order by idx to get the correct order of output artifacts from an operator.
	getOperatorToArtifactEdgesQuery := fmt.Sprintf(
		"SELECT %s FROM workflow_dag_edge WHERE workflow_dag_id = $1 AND type = '%s' ORDER BY idx;",
		allColumns(),
		OperatorToArtifactType,
	)

	var workflowDagEdges []WorkflowDagEdge
	err := db.Query(ctx, &workflowDagEdges, getOperatorToArtifactEdgesQuery, workflowDagId)
	return workflowDagEdges, err
}

func (r *standardReaderImpl) GetArtifactToOperatorEdges(
	ctx context.Context,
	workflowDagId uuid.UUID,
	db database.Database,
) ([]WorkflowDagEdge, error) {
	// We need to order by idx to get the correct order of input artifacts to an operator.
	getArtifactToOperatorEdgesQuery := fmt.Sprintf(
		"SELECT %s FROM workflow_dag_edge WHERE workflow_dag_id = $1 AND type = '%s' ORDER BY idx;",
		allColumns(),
		ArtifactToOperatorType,
	)

	var workflowDagEdges []WorkflowDagEdge
	err := db.Query(ctx, &workflowDagEdges, getArtifactToOperatorEdgesQuery, workflowDagId)
	return workflowDagEdges, err
}

func (r *standardReaderImpl) GetEdgesByWorkflowDagIds(
	ctx context.Context,
	workflowDagIds []uuid.UUID,
	db database.Database,
) ([]WorkflowDagEdge, error) {
	if len(workflowDagIds) == 0 {
		return nil, errors.New("Provided empty IDs list.")
	}

	query := fmt.Sprintf(
		"SELECT %s FROM workflow_dag_edge WHERE workflow_dag_id IN (%s);",
		allColumns(),
		stmt_preparers.GenerateArgsList(len(workflowDagIds), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(workflowDagIds)

	var workflowDagEdges []WorkflowDagEdge
	err := db.Query(ctx, &workflowDagEdges, query, args...)
	return workflowDagEdges, err
}

func (w *standardWriterImpl) DeleteEdgesByWorkflowDagIds(
	ctx context.Context,
	workflowDagIds []uuid.UUID,
	db database.Database,
) error {
	if len(workflowDagIds) == 0 {
		return nil
	}

	deleteStmt := fmt.Sprintf(
		"DELETE FROM workflow_dag_edge WHERE workflow_dag_id IN (%s);",
		stmt_preparers.GenerateArgsList(len(workflowDagIds), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(workflowDagIds)
	return db.Execute(ctx, deleteStmt, args...)
}
