package workflow

import (
	"context"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type Workflow struct {
	Id              uuid.UUID       `db:"id" json:"id"`
	UserId          uuid.UUID       `db:"user_id" json:"user_id"`
	Name            string          `db:"name" json:"name"`
	Description     string          `db:"description" json:"description"`
	Schedule        Schedule        `db:"schedule" json:"schedule"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	RetentionPolicy RetentionPolicy `db:"retention_policy" json:"retention_policy"`
}

type latestWorkflowResponse struct {
	Id          uuid.UUID              `db:"id" json:"id"`
	Name        string                 `db:"name" json:"name"`
	Description string                 `db:"description" json:"description"`
	CreatedAt   time.Time              `db:"created_at" json:"created_at"`
	LastRunAt   time.Time              `db:"last_run_at" json:"last_run_at"`
	Status      shared.ExecutionStatus `db:"status" json:"status"`
}

// Use to associate a workflow.name, workflow.id with workflow_dag_result.id (ENG-625)
type NotificationWorkflowMetadata struct {
	Id          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	DagResultId uuid.UUID `db:"dag_result_id" json:"dag_result_id"`
}

type WorkflowWatcherInfo struct {
	WorkflowId uuid.UUID `db:"workflow_id" json:"workflow_id"`
	Auth0Id    string    `db:"auth0_id" json:"auth0_id"`
}

type Reader interface {
	Exists(ctx context.Context, id uuid.UUID, db database.Database) (bool, error)
	GetWorkflow(ctx context.Context, id uuid.UUID, db database.Database) (*Workflow, error)
	GetWorkflowByWorkflowDagId(
		ctx context.Context,
		workflowDagId uuid.UUID,
		db database.Database,
	) (*Workflow, error)
	GetWorkflows(ctx context.Context, ids []uuid.UUID, db database.Database) ([]Workflow, error)
	GetWorkflowsByUser(
		ctx context.Context,
		userId uuid.UUID,
		db database.Database,
	) ([]Workflow, error)
	GetWorkflowByName(
		ctx context.Context,
		userId uuid.UUID,
		name string,
		db database.Database,
	) (*Workflow, error)
	GetAllWorkflows(
		ctx context.Context,
		db database.Database,
	) ([]Workflow, error)
	ValidateWorkflowOwnership(
		ctx context.Context,
		id uuid.UUID,
		organizationId string,
		db database.Database,
	) (bool, error)
	GetWorkflowsWithLatestRunResult(
		ctx context.Context,
		organizationId string,
		db database.Database,
	) ([]latestWorkflowResponse, error)
	GetNotificationWorkflowMetadata(
		ctx context.Context,
		ids []uuid.UUID,
		db database.Database,
	) (map[uuid.UUID]NotificationWorkflowMetadata, error)
	GetWatchersInBatch(
		ctx context.Context,
		workflowIds []uuid.UUID,
		db database.Database,
	) ([]WorkflowWatcherInfo, error)
}

type Writer interface {
	CreateWorkflow(
		ctx context.Context,
		userId uuid.UUID,
		name string,
		description string,
		schedule *Schedule,
		retentionPolicy *RetentionPolicy,
		db database.Database,
	) (*Workflow, error)
	UpdateWorkflow(
		ctx context.Context,
		id uuid.UUID, changes map[string]interface{},
		db database.Database,
	) (*Workflow, error)
	DeleteWorkflow(ctx context.Context, id uuid.UUID, db database.Database) error
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
