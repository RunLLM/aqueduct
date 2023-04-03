package storage_migration

import (
	"context"
	"time"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
)

/*
ListStorageMigrations returns a list of storage migration entries, ordered in reverse chronological order
(latest migrations first).

There are optional filters you can apply on the results:

		`status`:
			 Filters to storage migrations with the given status. Defaults to all statuses.
		`limit`:
			 The limit on the number of storage migrations returned. Defaults to all of them.
	         We always return storage migrations in descending chronological order, so setting this
		  	 to 1 will return the most recent storage migration.
		`completed-since`:
			  Unix timestamp. If set, we wil only return storage migrations that have completed since this time.
*/
func ListStorageMigrations(
	ctx context.Context,
	status *string,
	limit int,
	completedSince *time.Time,
	storageMigrationRepo repos.StorageMigration,
	DB database.Database,
) ([]models.StorageMigration, error) {
	migrations, err := storageMigrationRepo.List(ctx, DB)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list storage migrations.")
	}

	// Apply the filters in the following order: status, completedSince, then limit.
	if status != nil {
		migrations = filterStorageMigrationsByStatus(migrations, *status)
	}

	if completedSince != nil {
		migrations = filterStorageMigrationsByCompletedSince(migrations, *completedSince)
	}

	if limit >= 0 {
		return migrations[:limit], nil
	}
	return migrations, nil
}

func filterStorageMigrationsByStatus(migrations []models.StorageMigration, status string) []models.StorageMigration {
	filtered := make([]models.StorageMigration, 0, len(migrations))
	for _, migration := range migrations {
		if string(migration.ExecState.Status) == status {
			filtered = append(filtered, migration)
		}
	}
	return filtered
}

func filterStorageMigrationsByCompletedSince(migrations []models.StorageMigration, completedSince time.Time) []models.StorageMigration {
	var filtered []models.StorageMigration
	for _, migration := range migrations {
		finishedAt := migration.ExecState.Timestamps.FinishedAt
		if finishedAt != nil && finishedAt.After(completedSince) {
			filtered = append(filtered, migration)
		}
	}
	return filtered
}
