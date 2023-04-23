package _00026_drop_integration_validated_column

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
)

func UpPostgres(ctx context.Context, db database.Database) error {
	return db.Execute(ctx, upPostgresScript)
}

func UpSqlite(ctx context.Context, db database.Database) error {
	return db.Execute(ctx, sqliteScript)
}

func DownPostgres(ctx context.Context, db database.Database) error {
	return db.Execute(ctx, downPostgresScript)
}
