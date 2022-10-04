package workflow_dag

import (
	"context"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type DBWorkflowDag struct {
	Id            uuid.UUID            `db:"id" json:"id"`
	WorkflowId    uuid.UUID            `db:"workflow_id" json:"workflow_id"`
	CreatedAt     time.Time            `db:"created_at" json:"created_at"`
	StorageConfig shared.StorageConfig `db:"storage_config" json:"storage_config"`
	EngineConfig  shared.EngineConfig  `db:"engine_config" json:"engine_config"`

	/* Field not stored in DB */
	Metadata  *workflow.Workflow                `json:"metadata"`
	Operators map[uuid.UUID]operator.DBOperator `json:"operators,omitempty"`
	Artifacts map[uuid.UUID]artifact.DBArtifact `json:"artifacts,omitempty"`
}

func (dag *DBWorkflowDag) GetOperatorByName(name string) *operator.DBOperator {
	for _, op := range dag.Operators {
		if op.Name == name {
			return &op
		}
	}
	return nil
}

type Reader interface {
	GetWorkflowDag(ctx context.Context, id uuid.UUID, db database.Database) (*DBWorkflowDag, error)
	GetWorkflowDags(ctx context.Context, ids []uuid.UUID, db database.Database) ([]DBWorkflowDag, error)
	GetLatestWorkflowDag(ctx context.Context, workflowId uuid.UUID, db database.Database) (*DBWorkflowDag, error)
	GetWorkflowDagsByWorkflowId(
		ctx context.Context,
		workflowId uuid.UUID,
		db database.Database,
	) ([]DBWorkflowDag, error)
	GetWorkflowDagByWorkflowDagResultId(
		ctx context.Context,
		workflowDagResultId uuid.UUID,
		db database.Database,
	) (*DBWorkflowDag, error)
	GetWorkflowDagsByOperatorId(
		ctx context.Context,
		operatorId uuid.UUID,
		db database.Database,
	) ([]DBWorkflowDag, error)
	GetWorkflowDagsMapByArtifactResultIds(
		ctx context.Context,
		artifactResultIds []uuid.UUID,
		db database.Database,
	) (map[uuid.UUID]DBWorkflowDag, error)
}

type Writer interface {
	CreateWorkflowDag(
		ctx context.Context,
		workflowId uuid.UUID,
		storageConfig *shared.StorageConfig,
		engineConfig *shared.EngineConfig,
		db database.Database,
	) (*DBWorkflowDag, error)
	UpdateWorkflowDag(
		ctx context.Context,
		id uuid.UUID,
		changes map[string]interface{},
		db database.Database,
	) (*DBWorkflowDag, error)
	DeleteWorkflowDag(ctx context.Context, id uuid.UUID, db database.Database) error
	DeleteWorkflowDags(ctx context.Context, ids []uuid.UUID, db database.Database) error
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
