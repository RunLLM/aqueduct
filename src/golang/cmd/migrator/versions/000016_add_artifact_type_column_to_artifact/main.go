package _000016_add_artifact_type_column_to_artifact

import (
	"context"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/database"
	log "github.com/sirupsen/logrus"
)

func UpPostgres(ctx context.Context, db database.Database) error {
	err := db.Execute(ctx, upPostgresAddColumn)
	if err != nil {
		return err
	}

	err = migrateArtifact(ctx, db)
	if err != nil {
		return err
	}

	return db.Execute(ctx, upPostgresDropColumn)
}

func UpSqlite(ctx context.Context, db database.Database) error {
	err := db.Execute(ctx, sqliteAddColumn)
	if err != nil {
		if !strings.Contains(err.Error(), "duplicate column name: type") {
			return err
		} else {
			log.Info("Already has column type, ignoring error and proceeding...")
		}
	}

	err = migrateArtifact(ctx, db)
	if err != nil {
		return err
	}

	return db.Execute(ctx, sqliteDropColumn)
}

func DownPostgres(ctx context.Context, db database.Database) error {
	return db.Execute(ctx, downPostgresScript)
}
