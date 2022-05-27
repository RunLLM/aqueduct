package workflow_dag_edge

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type WorkflowDagEdge struct {
	WorkflowDagId uuid.UUID `db:"workflow_dag_id" json:"workflow_dag_id"`
	Type          Type      `db:"type" json:"type"`
	FromId        uuid.UUID `db:"from_id" json:"from_id"`
	ToId          uuid.UUID `db:"to_id" json:"to_id"`
	Idx           int16     `db:"idx" json:"idx"`
}

type Reader interface {
	GetOperatorToArtifactEdges(
		ctx context.Context,
		workflowDagId uuid.UUID,
		db database.Database,
	) ([]WorkflowDagEdge, error)
	GetArtifactToOperatorEdges(
		ctx context.Context,
		workflowDagId uuid.UUID,
		db database.Database,
	) ([]WorkflowDagEdge, error)
	GetEdgesByWorkflowDagIds(
		ctx context.Context,
		workflowDagIds []uuid.UUID,
		db database.Database,
	) ([]WorkflowDagEdge, error)
}

type Writer interface {
	CreateWorkflowDagEdge(
		ctx context.Context,
		workflowDagId uuid.UUID,
		edgeType Type,
		fromId uuid.UUID,
		toId uuid.UUID,
		idx int16,
		db database.Database,
	) (*WorkflowDagEdge, error)
	DeleteEdgesByWorkflowDagIds(
		ctx context.Context,
		workflowDagIds []uuid.UUID,
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
