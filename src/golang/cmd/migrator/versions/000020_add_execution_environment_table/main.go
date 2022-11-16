package _000020_add_execution_environment_table

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
)

func UpPostgres(ctx context.Context, db database.Database) error {
	return db.Execute(ctx, upPostgresScript)
}

func UpSqlite(ctx context.Context, db database.Database) error {
	return db.Execute(ctx, upSqliteScript)
}

func DownPostgres(ctx context.Context, db database.Database) error {
	return db.Execute(ctx, downPostgresScript)
}
