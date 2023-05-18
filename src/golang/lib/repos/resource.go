package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

// Resource defines all of the database operations that can be performed for a Resource.
type Resource interface {
	resourceReader
	resourceWriter
}

type resourceReader interface {
	// Get returns the Resource with ID.
	Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.Resource, error)

	// GetBatch returns the Resources with IDs.
	GetBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]models.Resource, error)

	// GetByConfigField returns the Resources with config fieldName=fieldValue.
	GetByConfigField(ctx context.Context, fieldName string, fieldValue string, DB database.Database) ([]models.Resource, error)

	// GetByNameAndUser returns the Resource named resourceName created by the user with the ID userID in the organization with the ID orgID.
	GetByNameAndUser(ctx context.Context, resourceName string, userID uuid.UUID, orgID string, DB database.Database) (*models.Resource, error)

	// GetByOrg returns the Resources by the organization with the ID orgID.
	GetByOrg(ctx context.Context, orgID string, DB database.Database) ([]models.Resource, error)

	// GetByServiceAndUser returns the Resources with the specified service created by the user with the ID userID.
	GetByServiceAndUser(ctx context.Context, service shared.Service, userID uuid.UUID, DB database.Database) ([]models.Resource, error)

	// GetByUser returns the Resources created by the org where the userID is equal to userID or it is NULL.
	GetByUser(ctx context.Context, orgID string, userID uuid.UUID, DB database.Database) ([]models.Resource, error)

	// ValidateOwnership checks whether the resource is owned by the user if the resource is of type userOnly otherwise, checks whether the resource is owned by the organizaion orgID.
	ValidateOwnership(ctx context.Context, resourceID uuid.UUID, orgID string, userID uuid.UUID, DB database.Database) (bool, error)
}

type resourceWriter interface {
	// Create inserts a new Resource with the specified fields.
	Create(
		ctx context.Context,
		orgID string,
		service shared.Service,
		name string,
		config *shared.ResourceConfig,
		DB database.Database,
	) (*models.Resource, error)

	// CreateForUser inserts a new Resource with the specified fields for the given user.
	CreateForUser(
		ctx context.Context,
		orgID string,
		userID uuid.UUID,
		service shared.Service,
		name string,
		config *shared.ResourceConfig,
		DB database.Database,
	) (*models.Resource, error)

	// Delete deletes the Resource with ID.
	Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error

	// Update applies changes to the Resource with ID. It returns the updated Resource.
	Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, DB database.Database) (*models.Resource, error)
}
