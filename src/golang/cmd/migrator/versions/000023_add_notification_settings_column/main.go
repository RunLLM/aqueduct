package _000023_add_notification_settings_column

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
