package sqlite

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type storageMigrationRepo struct {
	storageMigrationReader
	storageMigrationWriter
}

type storageMigrationReader struct{}

type storageMigrationWriter struct{}

func NewStorageMigrationRepo() repos.StorageMigration {
	return &storageMigrationRepo{
		storageMigrationReader: storageMigrationReader{},
		storageMigrationWriter: storageMigrationWriter{},
	}
}

// List returns all the storage migration entries in reverse chronological order, by creation time.
func (*storageMigrationReader) List(
	ctx context.Context,
	DB database.Database,
) ([]models.StorageMigration, error) {
	query := `SELECT * FROM storage_migration ORDER BY execution_state->>'timestamps'->>'registered_at' DESC`
	return getStorageMigrations(ctx, DB, query+";")
}

func (*storageMigrationWriter) Create(
	ctx context.Context,
	destIntegrationID *uuid.UUID,
	DB database.Database,
) (*models.StorageMigration, error) {
	cols := []string{
		models.StorageMigrationID,
		models.StorageMigrationDestIntegrationID,
		models.StorageMigrationExecutionState,
		models.StorageMigrationCurrent,
	}

	query := DB.PrepareInsertWithReturnAllStmt(models.StorageMigrationTable, cols, models.StorageMigrationCols())

	id, err := GenerateUniqueUUID(ctx, models.StorageMigrationTable, DB)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		id,
		destIntegrationID,
		createPendingExecState(),
		false, // current
	}

	return getStorageMigration(ctx, DB, query, args...)
}

// Returns nil if there is no current migration in the table.
func (*storageMigrationReader) Current(ctx context.Context, DB database.Database) (*models.StorageMigration, error) {
	query := `SELECT * FROM storage_migration WHERE current = true;`
	return getStorageMigration(ctx, DB, query)
}

func (*storageMigrationWriter) Update(ctx context.Context, id uuid.UUID, changes map[string]interface{}, DB database.Database) (*models.StorageMigration, error) {
	var storageMigration models.StorageMigration
	err := repos.UpdateRecordToDest(
		ctx,
		&storageMigration,
		changes,
		models.StorageMigrationTable,
		models.StorageMigrationID,
		id,
		models.StorageMigrationCols(),
		DB,
	)
	return &storageMigration, err
}

func getStorageMigrations(ctx context.Context, DB database.Database, query string, args ...interface{}) ([]models.StorageMigration, error) {
	var storageMigrations []models.StorageMigration
	err := DB.Query(ctx, &storageMigrations, query, args...)
	return storageMigrations, err
}

func getStorageMigration(
	ctx context.Context,
	DB database.Database,
	query string,
	args ...interface{},
) (*models.StorageMigration, error) {
	storageMigrations, err := getStorageMigrations(ctx, DB, query, args...)
	if err != nil {
		return nil, err
	}

	if len(storageMigrations) == 0 {
		return nil, database.ErrNoRows()
	}

	if len(storageMigrations) > 1 {
		return nil, errors.Newf("Expected 1 storage migration entry but got %v", len(storageMigrations))
	}

	return &storageMigrations[0], err
}
