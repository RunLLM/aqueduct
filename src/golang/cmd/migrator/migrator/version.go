package migrator

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/repos/sqlite"
	"github.com/aqueducthq/aqueduct/lib/database"
)

// Version returns the current schema version of the database db
// and whether the schema version is dirty.
// It also returns an error, if any.
func Version(ctx context.Context, db database.Database) (int64, bool, error) {
	schemaVersion, err := sqlite.NewSchemaVersionRepo().GetCurrent(ctx, db)
	if err != nil {
		return -1, false, err
	}

	return schemaVersion.Version, schemaVersion.Dirty, nil
}
