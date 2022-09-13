package _000018_add_dag_result_exec_state_column

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
)

func UpPostgres(ctx context.Context, db database.Database) error {
	err := db.Execute(ctx, postgresAddColScript)
	if err != nil {
		return err
	}

	return backfill(ctx, db)
}

func UpSqlite(ctx context.Context, db database.Database) error {
	err := db.Execute(ctx, sqliteAddColScript)
	if err != nil {
		return err
	}

	return backfill(ctx, db)
}

func DownPostgres(ctx context.Context, db database.Database) error {
	return db.Execute(ctx, downPostgresScript)
}
