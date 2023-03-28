package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/google/uuid"
)

type StorageMigration interface {
	storageMigrationReader
	storageMigrationWriter
}

type storageMigrationReader interface {
	// GetCurrent returns the ID of the current integration that is being used for storage.
	GetCurrent(ctx context.Context) (uuid.UUID, error)
}

type storageMigrationWriter interface {
	// Create inserts a new storage migration entry with all the starter fields.
	// A nil integration id refers to the local filesystem.
	Create(
		ctx context.Context,
		destIntegrationID *uuid.UUID,
		DB database.Database,
	) (*models.StorageMigration, error)

	// Update updates the storage migration entry with the given ID.
	Update(
		ctx context.Context,
		id uuid.UUID,
		changes map[string]interface{},
		DB database.Database,
	) (*models.StorageMigration, error)
}
