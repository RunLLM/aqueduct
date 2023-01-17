package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
)

// SchemaVersion defines all of the database operations that can be performed for an SchemaVersion.
type SchemaVersion interface {
	schemaVersionReader
	schemaVersionWriter
}

type schemaVersionReader interface {
	// Get returns the SchemaVersion by the version number.
	Get(ctx context.Context, version int64, DB database.Database) (*models.SchemaVersion, error)

	// GetCurrent returns the current SchemaVersion.
	GetCurrent(ctx context.Context, DB database.Database) (*models.SchemaVersion, error)
}

type schemaVersionWriter interface {
	// Create inserts a new SchemaVersion with the specified fields.
	Create(
		ctx context.Context,
		version int64,
		name string,
		DB database.Database,
	) (*models.SchemaVersion, error)

	// Delete deletes the SchemaVersion by the version number.
	Delete(ctx context.Context, version int64, DB database.Database) error

	// Update applies changes to the SchemaVersion with the version number. It returns the updated SchemaVersion.
	Update(ctx context.Context, version int64, changes map[string]interface{}, DB database.Database) (*models.SchemaVersion, error)
}
