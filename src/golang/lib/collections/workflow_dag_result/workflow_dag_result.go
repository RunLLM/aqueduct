package workflow_dag_result

import (
	"context"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/user"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type WorkflowDagResult struct {
	Id            uuid.UUID              `db:"id" json:"id"`
	WorkflowDagId uuid.UUID              `db:"workflow_dag_id" json:"workflow_dag_id"`
	Status        shared.ExecutionStatus `db:"status" json:"status"`
	CreatedAt     time.Time              `db:"created_at" json:"created_at"`
}

type Reader interface {
	GetWorkflowDagResult(
		ctx context.Context,
		id uuid.UUID,
		db database.Database,
	) (*WorkflowDagResult, error)
	GetWorkflowDagResults(
		ctx context.Context,
		ids []uuid.UUID,
		db database.Database,
	) ([]WorkflowDagResult, error)
	GetWorkflowDagResultsByWorkflowId(
		ctx context.Context,
		workflowId uuid.UUID,
		db database.Database,
	) ([]WorkflowDagResult, error)
	GetKOffsetWorkflowDagResultsByWorkflowId(
		ctx context.Context,
		workflowId uuid.UUID,
		k int,
		db database.Database,
	) ([]WorkflowDagResult, error)
}

type Writer interface {
	CreateWorkflowDagResult(
		ctx context.Context,
		workflowDagId uuid.UUID,
		db database.Database,
	) (*WorkflowDagResult, error)
	UpdateWorkflowDagResult(
		ctx context.Context,
		id uuid.UUID,
		changes map[string]interface{},
		workflowReader workflow.Reader,
		notificationWriter notification.Writer,
		userReader user.Reader,
		db database.Database,
	) (*WorkflowDagResult, error)
	DeleteWorkflowDagResult(ctx context.Context, id uuid.UUID, db database.Database) error
	DeleteWorkflowDagResults(ctx context.Context, ids []uuid.UUID, db database.Database) error
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
