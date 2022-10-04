package workflow_dag_result

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type sqliteReaderImpl struct {
	standardReaderImpl
}

type sqliteWriterImpl struct {
	standardWriterImpl
}

func newSqliteReader() Reader {
	return &sqliteReaderImpl{standardReaderImpl{}}
}

func newSqliteWriter() Writer {
	return &sqliteWriterImpl{standardWriterImpl{}}
}

func (r *sqliteReaderImpl) GetKOffsetWorkflowDagResultsByWorkflowId(
	ctx context.Context,
	workflowId uuid.UUID,
	k int,
	db database.Database,
) ([]WorkflowDagResult, error) {
	// Get all workflow DAGs for the workflow specified by `workflowId` except for the k latest.
	// https://itecnote.com/tecnote/sqlite-limit-offset-query/
	// `LIMIT <skip>, <count>` is equivalent to `LIMIT <count> OFFSET <skip>`
	query := fmt.Sprintf(`
		SELECT %s FROM workflow_dag_result, workflow_dag 
		WHERE workflow_dag_result.workflow_dag_id = workflow_dag.id AND workflow_dag.workflow_id = $1
		ORDER BY workflow_dag_result.created_at DESC
		LIMIT $2, -1;`,
		allColumnsWithPrefix())

	var workflowDagResults []WorkflowDagResult
	err := db.Query(ctx, &workflowDagResults, query, workflowId, k)
	return workflowDagResults, err
}

func (w *sqliteWriterImpl) CreateWorkflowDagResult(
	ctx context.Context,
	workflowDagId uuid.UUID,
	execState *shared.ExecutionState,
	db database.Database,
) (*WorkflowDagResult, error) {
	insertColumns := []string{IdColumn, WorkflowDagIdColumn, StatusColumn, CreatedAtColumn, ExecStateColumn}
	insertWorkflowDagResultStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	id, err := utils.GenerateUniqueUUID(ctx, tableName, db)
	if err != nil {
		return nil, err
	}

	if execState.Timestamps == nil || execState.Timestamps.PendingAt == nil {
		return nil, ErrInvalidPendingTimestamp
	}

	args := []interface{}{id, workflowDagId, execState.Status, *(execState.Timestamps.PendingAt), execState}

	var workflowDagResult WorkflowDagResult
	err = db.Query(ctx, &workflowDagResult, insertWorkflowDagResultStmt, args...)
	return &workflowDagResult, err
}
