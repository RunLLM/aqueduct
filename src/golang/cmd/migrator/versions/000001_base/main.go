package _000001_base

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
)

func UpPostgres(ctx context.Context, db database.Database) error {
	return db.Execute(ctx, upPostgresScript)
}

func UpSqlite(ctx context.Context, db database.Database) error {
	return db.Execute(ctx, sqliteScript)
}

func DownPostgres(ctx context.Context, db database.Database) error {
	return errors.New("Migrating down from version 000001 is not allowed.")
}
