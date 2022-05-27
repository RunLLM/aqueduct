package workflow_watcher

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type WorkflowWatcher struct {
	WorkflowId uuid.UUID `db:"workflow_id"`
	UserId     uuid.UUID `db:"user_id"`
}

type Reader interface {
	GetWorkflowWatcher(
		ctx context.Context,
		workflowId uuid.UUID,
		userId uuid.UUID,
		db database.Database,
	) (*WorkflowWatcher, error)
	GetWorkflowWatchersByWorkflow(
		ctx context.Context,
		workflowId uuid.UUID,
		db database.Database,
	) ([]WorkflowWatcher, error)
}

type Writer interface {
	CreateWorkflowWatcher(
		ctx context.Context,
		workflowId uuid.UUID,
		userId uuid.UUID,
		db database.Database,
	) (*WorkflowWatcher, error)
	DeleteWorkflowWatcher(
		ctx context.Context,
		workflowId uuid.UUID,
		userId uuid.UUID,
		db database.Database,
	) error
	DeleteWorkflowWatcherByWorkflowId(
		ctx context.Context,
		workflowId uuid.UUID,
		db database.Database,
	) error
	DeleteWorkflowWatchersByUser(
		ctx context.Context,
		userId uuid.UUID,
		db database.Database,
	) error
}

func NewReader(dbConf *database.DatabaseConfig) (Reader, error) {
	if dbConf.Type == database.PostgresType {
		return newPostgresReader(), nil
	}

	if dbConf.Type == database.SqliteType {
		return newSqliteReader(), nil
	}

	return nil, database.ErrUnsupportedDbType
}

func NewWriter(dbConf *database.DatabaseConfig) (Writer, error) {
	if dbConf.Type == database.PostgresType {
		return newPostgresWriter(), nil
	}

	if dbConf.Type == database.SqliteType {
		return newSqliteWriter(), nil
	}

	return nil, database.ErrUnsupportedDbType
}
