package migrator

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/schema_version"
	"github.com/aqueducthq/aqueduct/lib/database"
)

// Version returns the current schema version of the database db
// and whether the schema version is dirty.
// It also returns an error, if any.
func Version(ctx context.Context, db database.Database) (int64, bool, error) {
	reader, err := schema_version.NewReader(db.Config())
	if err != nil {
		return -1, false, err
	}

	schemaVersion, err := reader.GetCurrentSchemaVersion(ctx, db)
	if err != nil {
		return -1, false, err
	}

	return schemaVersion.Version, schemaVersion.Dirty, nil
}
