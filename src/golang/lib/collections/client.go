package collections

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
)

// RequireSchemaVersion returns an error if the database schema version is
// not at least `version`.
func RequireSchemaVersion(
	ctx context.Context,
	version int64,
	schemaVersionRepo repos.SchemaVersion,
	db database.Database,
) error {
	currentVersion, err := schemaVersionRepo.GetCurrent(ctx, db)
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
