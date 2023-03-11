package migrator

import (
	"context"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	log "github.com/sirupsen/logrus"
)

// GoTo performs a migration to version. It returns an error, if any.
func GoTo(ctx context.Context, version int64, db database.Database) error {
	current, dirty, err := Version(ctx, db)
	if err != nil {
		if strings.Contains(err.Error(), database.ErrCodeTableDoesNotExist) ||
			strings.Contains(err.Error(), "no such table") {
			// We are running a schema migration for the first time, so
			// the schema_version table does not exist.
			current, dirty = 0, false
		} else {
			return errors.Wrap(err, "Unable to check current version.")
		}
	}

	if current == version && !dirty {
		log.Infof("The schema version is already %d.", version)
		return nil
	}

	if version < current && db.Type() == database.SqliteType {
		return errors.New("Down schema changes are not supported for SQLite.")
	}

	migrator, err := newMigrator(current, version, dirty)
	if err != nil {
		return errors.Wrap(err, "Unable to initiate migration.")
	}

	if err := migrator.execute(ctx, db); err != nil {
		return errors.Wrap(err, "Schema migration was unsuccessful.")
	}

	return nil
}

// Up performs a single up migration. It returns an error, if any.
func Up(ctx context.Context, db database.Database) error {
	current, dirty, err := Version(ctx, db)
	if err != nil {
		return errors.Wrap(err, "Unable to check current version.")
	}

	migrator, err := newMigrator(current, current+1, dirty)
	if err != nil {
		return errors.Wrap(err, "Unable to initiate migration.")
	}

	if err := migrator.execute(ctx, db); err != nil {
		return errors.Wrap(err, "Schema migration was unsuccessful.")
	}

	return nil
}

// Down performs a single down migration. It returns an error, if any.
func Down(ctx context.Context, db database.Database) error {
	if db.Type() == database.SqliteType {
		return errors.New("Down schema changes are not supported for SQLite.")
	}

	current, dirty, err := Version(ctx, db)
	if err != nil {
		return errors.Wrap(err, "Unable to check current version.")
	}

	migrator, err := newMigrator(current, current-1, dirty)
	if err != nil {
		return errors.Wrap(err, "Unable to initiate migration.")
	}

	if err := migrator.execute(ctx, db); err != nil {
		return errors.Wrap(err, "Schema migration was unsuccessful.")
	}

	return nil
}
