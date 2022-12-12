package sqlite

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/google/uuid"
)

type watcherRepo struct {
	watcherReader
	watcherWriter
}

type watcherReader struct{}

type watcherWriter struct{}

func NewWatcherRepo() repos.Watcher {
	return &watcherRepo{
		watcherReader: watcherReader{},
		watcherWriter: watcherWriter{},
	}
}

func (*watcherWriter) Create(
	ctx context.Context,
	workflowID uuid.UUID,
	userID uuid.UUID,
	DB database.Database,
) (*models.Watcher, error) {
	cols := []string{
		models.WatcherWorkflowID,
		models.WatcherUserID,
	}
	query := DB.PrepareInsertWithReturnAllStmt(models.WatcherTable, cols, models.WatcherCols())

	args := []interface{}{
		workflowID,
		userID,
	}

	var watcher models.Watcher
	err := DB.Query(ctx, &watcher, query, args...)
	return &watcher, err
}

func (*watcherWriter) Delete(
	ctx context.Context,
	workflowID uuid.UUID,
	userID uuid.UUID,
	DB database.Database,
) error {
	query := `DELETE FROM workflow_watcher
	WHERE (workflow_id, user_id) = ($1, $2);`
	args := []interface{}{workflowID, userID}

	return DB.Execute(ctx, query, args...)
}

func (*watcherWriter) DeleteByWorkflow(ctx context.Context, workflowID uuid.UUID, DB database.Database) error {
	query := `DELETE FROM workflow_watcher
	WHERE workflow_id = $1;`
	args := []interface{}{workflowID}

	return DB.Execute(ctx, query, args...)
}
