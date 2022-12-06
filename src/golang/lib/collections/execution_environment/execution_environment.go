package execution_environment

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type DBExecutionEnvironment struct {
	Id               uuid.UUID `db:"id"`
	Spec             Spec      `db:"spec"`
	Hash             uuid.UUID `db:"hash"`
	GarbageCollected bool      `db:"garbage_collected"`
}

type Reader interface {
	GetExecutionEnvironment(ctx context.Context, id uuid.UUID, db database.Database) (*DBExecutionEnvironment, error)
	GetExecutionEnvironments(ctx context.Context, ids []uuid.UUID, db database.Database) ([]DBExecutionEnvironment, error)
	GetActiveExecutionEnvironmentByHash(ctx context.Context, hash uuid.UUID, db database.Database) (*DBExecutionEnvironment, error)
	GetActiveExecutionEnvironmentsByOperatorID(
		ctx context.Context,
		opIDs []uuid.UUID,
		db database.Database,
	) (map[uuid.UUID]DBExecutionEnvironment, error)
	GetUnusedExecutionEnvironments(
		ctx context.Context,
		db database.Database,
	) ([]DBExecutionEnvironment, error)
}

type Writer interface {
	CreateExecutionEnvironment(
		ctx context.Context,
		spec *Spec,
		hash uuid.UUID,
		db database.Database,
	) (*DBExecutionEnvironment, error)
	UpdateExecutionEnvironment(
		ctx context.Context,
		id uuid.UUID,
		changes map[string]interface{},
		db database.Database,
	) (*DBExecutionEnvironment, error)
	DeleteExecutionEnvironment(ctx context.Context, id uuid.UUID, db database.Database) error
	DeleteExecutionEnvironments(ctx context.Context, ids []uuid.UUID, db database.Database) error
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
