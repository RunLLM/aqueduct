package migrator

import (
	"context"
	"math"

	"github.com/aqueducthq/aqueduct/lib/collections/schema_version"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
)

type direction string

const (
	upDirection   direction = "up"
	downDirection direction = "down"
)

type migrationStep struct {
	*migration
	version int64
}

type migrator struct {
	dir   direction
	steps []*migrationStep
}

// newMigrator creates a migrator to go from currentVersion to toVersion.
// It takes into account whether the current version is dirty.
// It also returns an error, if any.
func newMigrator(currentVersion, toVersion int64, dirty bool) (*migrator, error) {
	dir := upDirection
	if toVersion < currentVersion {
		dir = downDirection
	}

	numSteps := int64(math.Abs(float64(currentVersion-toVersion))) + 1
	steps := make([]*migrationStep, 0, numSteps)

	version := currentVersion
	if dir == upDirection && !dirty {
		// For up migrations, if the current version is not dirty, we should start at the next version.
		// If the current version is dirty, it needs to be repeated.
		version += 1
	}

	for {
		m, ok := registeredMigrations[version]
		if !ok {
			return nil, errors.Newf("Schema migration version %d has not been registered.", version)
		}

		steps = append(steps, &migrationStep{
			migration: m,
			version:   version,
		})

		if dir == upDirection {
			if version == toVersion {
				// This is the last step for an up migration.
				break
			}
			version += 1
		} else {
			if version-1 == toVersion {
				// This is the last step for a down migration.
				break
			}
			version -= 1
		}
	}

	return &migrator{
		dir:   dir,
		steps: steps,
	}, nil
}

// execute performs all the migration steps for m.
// It returns an error, if any.
func (m *migrator) execute(ctx context.Context, db database.Database) error {
	for _, step := range m.steps {
		if m.dir == upDirection {
			if err := executeUp(ctx, db, step); err != nil {
				return err
			}
		} else {
			if err := executeDown(ctx, db, step); err != nil {
				return err
			}
		}
	}

	return nil
}

func executeUp(ctx context.Context, db database.Database, step *migrationStep) error {
	log.Infof("Starting migrate up to schema version %d.", step.version)

	if step.version > 1 {
		// schema_version table does not exist before version 1
		if err := createSchemaVersionRecord(ctx, step.version, step.name, db); err != nil {
			return errors.Wrap(err, "Unable to create schema version record.")
		}
	}

	var up migrationFunc
	switch db.Type() {
	case database.PostgresType:
		up = step.upPostgres
	case database.SqliteType:
		up = step.upSqlite
	default:
		return errors.Newf("Unknown database type: %v", db.Type())
	}

	err := up(ctx, db)
	if err != nil {
		log.Errorf("Failed migrate up to schema version %d.", step.version)
		return err
	}

	if err := setSchemaVersionRecordDirty(ctx, step.version, false /* dirty */, db); err != nil {
		return errors.Wrap(err, "Unable to set schema version record to not dirty.")
	}

	log.Infof("Completed migrate up to schema version %d.", step.version)
	return nil
}

func executeDown(ctx context.Context, db database.Database, step *migrationStep) error {
	log.Infof("Starting migrate down to schema version %d.", step.version)

	if err := setSchemaVersionRecordDirty(ctx, step.version, true /* dirty */, db); err != nil {
		return errors.Wrap(err, "Unable to set schema version record to dirty.")
	}

	if db.Type() != database.PostgresType {
		return errors.Newf("Down schema changes are not defined for: %v", db.Type())
	}

	err := step.downPostgres(ctx, db)
	if err != nil {
		log.Errorf("Failed migrate down to schema version %d.", step.version)
		return err
	}

	if err := deleteSchemaVersionRecord(ctx, step.version, db); err != nil {
		return errors.Wrap(err, "Unable to delete schema version record.")
	}

	log.Infof("Completed migrate down to schema version %d.", step.version)
	return nil
}

// createSchemaVersionRecord creates a new record in the database for the schema version.
// If the schema version record already exists (due to a previous dirty run), then the
// schema version is simply set to dirty.
func createSchemaVersionRecord(ctx context.Context, version int64, name string, db database.Database) error {
	schemaVersionRepo := sqlite.NewSchemaVersionRepo()

	writer, err := schema_version.NewWriter(db.Config())
	if err != nil {
		return err
	}

	_, err = schemaVersionRepo.Get(ctx, version, db)
	if err == nil {
		// The schema version record already exists from a previous migration, so we just set dirty to true.
		return setSchemaVersionRecordDirty(ctx, version, true /* dirty */, db)
	}

	_, err = schemaVersionRepo.Create(ctx, version, name, db)
	return err
}

// setSchemaVersionRecordDirty updates the dirty column in the database for the schema version.
func setSchemaVersionRecordDirty(ctx context.Context, version int64, dirty bool, db database.Database) error {
	writer, err := schema_version.NewWriter(db.Config())
	if err != nil {
		return err
	}

	_, err = writer.UpdateSchemaVersion(
		ctx,
		version,
		map[string]interface{}{
			schema_version.DirtyColumn: dirty,
		},
		db,
	)
	return err
}

// deleteSchemaVersionRecord deletes the record for the schema version from the database.
func deleteSchemaVersionRecord(ctx context.Context, version int64, db database.Database) error {
	writer, err := schema_version.NewWriter(db.Config())
	if err != nil {
		return err
	}

	return writer.DeleteSchemaVersion(ctx, version, db)
}
