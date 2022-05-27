package workflow_watcher

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type standardReaderImpl struct{}

type standardWriterImpl struct{}

func (w *standardWriterImpl) CreateWorkflowWatcher(
	ctx context.Context,
	workflowId uuid.UUID,
	userId uuid.UUID,
	db database.Database,
) (*WorkflowWatcher, error) {
	insertColumns := []string{
		WorkflowIdColumn, UserIdColumn,
	}
	insertWorkflowWatcherStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	args := []interface{}{
		workflowId, userId,
	}

	var workflowWatcher WorkflowWatcher
	err := db.Query(ctx, &workflowWatcher, insertWorkflowWatcherStmt, args...)
	return &workflowWatcher, err
}

func (r *standardReaderImpl) GetWorkflowWatcher(
	ctx context.Context,
	workflowId uuid.UUID,
	userId uuid.UUID,
	db database.Database,
) (*WorkflowWatcher, error) {
	getWorkflowWatcherQuery := fmt.Sprintf(
		"SELECT %s FROM workflow_watcher WHERE (workflow_id, user_id) = ($1, $2);",
		allColumns(),
	)
	var workflowWatcher WorkflowWatcher

	err := db.Query(ctx, &workflowWatcher, getWorkflowWatcherQuery, workflowId, userId)
	return &workflowWatcher, err
}

func (r *standardReaderImpl) GetWorkflowWatchersByWorkflow(
	ctx context.Context,
	workflowId uuid.UUID,
	db database.Database,
) ([]WorkflowWatcher, error) {
	getWorkflowWatchersQuery := fmt.Sprintf(
		"SELECT %s FROM workflow_watcher WHERE workflow_id = $1;",
		allColumns(),
	)
	var workflowWatchers []WorkflowWatcher

	err := db.Query(ctx, &workflowWatchers, getWorkflowWatchersQuery, workflowId)
	return workflowWatchers, err
}

func (w *standardWriterImpl) DeleteWorkflowWatcher(
	ctx context.Context,
	workflowId uuid.UUID,
	userId uuid.UUID,
	db database.Database,
) error {
	deleteWorkflowWatcherStmt := `DELETE FROM workflow_watcher WHERE (workflow_id, user_id) = ($1, $2);`
	return db.Execute(ctx, deleteWorkflowWatcherStmt, workflowId, userId)
}

func (w *standardWriterImpl) DeleteWorkflowWatcherByWorkflowId(
	ctx context.Context,
	workflowId uuid.UUID,
	db database.Database,
) error {
	deleteWorkflowWatcherStmt := `DELETE FROM workflow_watcher WHERE workflow_id = $1;`
	return db.Execute(ctx, deleteWorkflowWatcherStmt, workflowId)
}

func (w *standardWriterImpl) DeleteWorkflowWatchersByUser(
	ctx context.Context,
	userId uuid.UUID,
	db database.Database,
) error {
	deleteWorkflowWatcherStmt := `DELETE FROM workflow_watcher WHERE user_id = $1;`
	return db.Execute(ctx, deleteWorkflowWatcherStmt, userId)
}
