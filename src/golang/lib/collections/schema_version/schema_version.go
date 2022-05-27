package schema_version

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
)

type SchemaVersion struct {
	Version int64  `db:"version"`
	Dirty   bool   `db:"dirty"`
	Name    string `db:"name"`
}

type Reader interface {
	GetSchemaVersion(ctx context.Context, version int64, db database.Database) (*SchemaVersion, error)
	GetCurrentSchemaVersion(ctx context.Context, db database.Database) (*SchemaVersion, error)
}

type Writer interface {
	CreateSchemaVersion(
		ctx context.Context,
		version int64,
		name string,
		db database.Database,
	) (*SchemaVersion, error)
	UpdateSchemaVersion(
		ctx context.Context,
		version int64,
		changes map[string]interface{},
		db database.Database,
	) (*SchemaVersion, error)
	DeleteSchemaVersion(ctx context.Context, version int64, db database.Database) error
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
