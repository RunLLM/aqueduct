package collections

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/schema_version"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/dropbox/godropbox/errors"
)

// RequireSchemaVersion returns an error if the database schema version is
// not at least `version`.
func RequireSchemaVersion(
	ctx context.Context,
	version int64,
	schemaVersionReader schema_version.Reader,
	db database.Database,
) error {
	currentVersion, err := schemaVersionReader.GetCurrentSchemaVersion(ctx, db)
	if err != nil {
		return err
	}

	if currentVersion.Version < version {
		return errors.Newf(
			"Current version is %d, but %d is required.",
			currentVersion.Version,
			version,
		)
	}
	return nil
}
