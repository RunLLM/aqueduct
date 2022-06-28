package operator_result

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type OperatorResult struct {
	Id                  uuid.UUID              `db:"id" json:"id"`
	WorkflowDagResultId uuid.UUID              `db:"workflow_dag_result_id" json:"workflow_dag_result_id"`
	OperatorId          uuid.UUID              `db:"operator_id" json:"operator_id"`
	Status              shared.ExecutionStatus `db:"status" json:"status"`
	Metadata            NullExecutionLogs      `db:"metadata" json:"metadata"`
}

type Reader interface {
	GetOperatorResult(ctx context.Context, id uuid.UUID, db database.Database) (*OperatorResult, error)
	GetOperatorResults(ctx context.Context, ids []uuid.UUID, db database.Database) ([]OperatorResult, error)
	GetOperatorResultByWorkflowDagResultIdAndOperatorId(
		ctx context.Context,
		workflowDagResultId,
		operatorId uuid.UUID,
		db database.Database,
	) (*OperatorResult, error)
	GetOperatorResultsByWorkflowDagResultIds(
		ctx context.Context,
		workflowDagResultIds []uuid.UUID,
		db database.Database,
	) ([]OperatorResult, error)
}

type Writer interface {
	CreateOperatorResult(
		ctx context.Context,
		workflowDagResultId uuid.UUID,
		operatorId uuid.UUID,
		db database.Database,
	) (*OperatorResult, error)
	UpdateOperatorResult(
		ctx context.Context,
		id uuid.UUID,
		changes map[string]interface{},
		db database.Database,
	) (*OperatorResult, error)
	DeleteOperatorResult(ctx context.Context, id uuid.UUID, db database.Database) error
	DeleteOperatorResults(ctx context.Context, ids []uuid.UUID, db database.Database) error
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
