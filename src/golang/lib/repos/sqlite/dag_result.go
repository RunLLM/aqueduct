package sqlite

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type dagResultRepo struct {
	dagResultReader
	dagResultWriter
}

type dagResultReader struct{}

type dagResultWriter struct{}

func NewDAGResultRepo() repos.DAGResult {
	return &dagResultRepo{
		dagResultReader: dagResultReader{},
		dagResultWriter: dagResultWriter{},
	}
}

func (*dagResultReader) Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.DAGResult, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM workflow_dag_result WHERE id = $1;`,
		models.DAGResultCols(),
	)
	args := []interface{}{ID}

	return getDAGResult(ctx, DB, query, args...)
}

func (*dagResultReader) GetBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]models.DAGResult, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM workflow_dag_result WHERE id in (%s);`,
		models.DAGResultCols(),
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)

	return getDAGResults(ctx, DB, query, args...)
}

func (*dagResultReader) GetByWorkflow(ctx context.Context, workflowID uuid.UUID, DB database.Database) ([]models.DAGResult, error) {
	query := fmt.Sprintf(
		`SELECT %s 
		FROM workflow_dag_result, workflow_dag 
		WHERE 
			workflow_dag_result.workflow_dag_id = workflow_dag.id 
			AND workflow_dag.workflow_id = $1;`,
		models.DAGResultColsWithPrefix(),
	)
	args := []interface{}{workflowID}

	return getDAGResults(ctx, DB, query, args...)
}

func (*dagResultReader) GetKOffsetByWorkflow(ctx context.Context, workflowID uuid.UUID, k int, DB database.Database) ([]models.DAGResult, error) {
	// https://itecnote.com/tecnote/sqlite-limit-offset-query/
	// `LIMIT <skip>, <count>` is equivalent to `LIMIT <count> OFFSET <skip>`
	query := fmt.Sprintf(
		`SELECT %s 
		FROM workflow_dag_result, workflow_dag 
		WHERE 
			workflow_dag_result.workflow_dag_id = workflow_dag.id 
			AND workflow_dag.workflow_id = $1
		ORDER BY workflow_dag_result.created_at DESC
		LIMIT $2, -1;`,
		models.DAGResultColsWithPrefix(),
	)
	args := []interface{}{workflowID, k}

	return getDAGResults(ctx, DB, query, args...)
}

func (*dagResultReader) GetWorkflowMetadataBatch(
	ctx context.Context,
	IDs []uuid.UUID,
	DB database.Database,
) (map[uuid.UUID]views.DAGResultWorkflowMetadata, error) {
	query := fmt.Sprintf(
		`SELECT 
			workflow.id, workflow.name, workflow_dag_result.id AS dag_result_id
		FROM 
			workflow, workflow_dag, workflow_dag_result 
		WHERE 
			workflow_dag_result.workflow_dag_id = workflow_dag.id 
			AND workflow.id = workflow_dag.workflow_id 
			AND workflow_dag_result.id IN (%s);`,
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)

	var workflowMetadata []views.DAGResultWorkflowMetadata
	if err := DB.Query(ctx, &workflowMetadata, query, args...); err != nil {
		return nil, err
	}

	dagResultToWorkflowMetadata := make(
		map[uuid.UUID]views.DAGResultWorkflowMetadata,
		len(workflowMetadata),
	)
	for _, metadata := range workflowMetadata {
		dagResultToWorkflowMetadata[metadata.DAGResultID] = metadata
	}

	return dagResultToWorkflowMetadata, nil
}

func (*dagResultWriter) Create(
	ctx context.Context,
	dagID uuid.UUID,
	execState *shared.ExecutionState,
	DB database.Database,
) (*models.DAGResult, error) {
	cols := []string{
		models.DAGResultID,
		models.DAGResultDagID,
		models.DAGResultStatus,
		models.DAGResultCreatedAt,
		models.DAGResultExecState,
	}
	query := DB.PrepareInsertWithReturnAllStmt(models.DAGResultTable, cols, models.DAGResultCols())

	ID, err := GenerateUniqueUUID(ctx, models.DAGResultTable, DB)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		ID,
		dagID,
		execState.Status,
		*(execState.Timestamps.PendingAt),
		execState,
	}

	return getDAGResult(ctx, DB, query, args...)
}

func (*dagResultWriter) Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error {
	return deleteDAGResults(ctx, DB, []uuid.UUID{ID})
}

func (*dagResultWriter) DeleteBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) error {
	return deleteDAGResults(ctx, DB, IDs)
}

func (*dagResultWriter) Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, DB database.Database) (*models.DAGResult, error) {
	var dagResult models.DAGResult
	err := repos.UpdateRecordToDest(
		ctx,
		&dagResult,
		changes,
		models.DAGResultTable,
		models.DAGResultID,
		ID,
		models.DAGResultCols(),
		DB,
	)

	return &dagResult, err
}

func getDAGResults(ctx context.Context, DB database.Database, query string, args ...interface{}) ([]models.DAGResult, error) {
	var dagResults []models.DAGResult
	err := DB.Query(ctx, &dagResults, query, args...)
	return dagResults, err
}

func getDAGResult(ctx context.Context, DB database.Database, query string, args ...interface{}) (*models.DAGResult, error) {
	dagResults, err := getDAGResults(ctx, DB, query, args...)
	if err != nil {
		return nil, err
	}

	if len(dagResults) == 0 {
		return nil, database.ErrNoRows()
	}

	if len(dagResults) != 1 {
		return nil, errors.Newf("Expected 1 DAGResult but got %v", len(dagResults))
	}

	return &dagResults[0], nil
}

func deleteDAGResults(ctx context.Context, DB database.Database, IDs []uuid.UUID) error {
	if len(IDs) == 0 {
		return nil
	}

	query := fmt.Sprintf(
		`DELETE FROM workflow_dag_result WHERE id IN (%s)`,
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)

	return DB.Execute(ctx, query, args...)
}
