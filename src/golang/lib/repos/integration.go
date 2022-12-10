package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/google/uuid"
)

// Integration defines all of the database operations that can be performed for a Integration.
type Integration interface {
	integrationReader
	integrationWriter
}

type integrationReader interface {
	// Get returns the Integration with ID.
	Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.Integration, error)

	// GetBatch returns the Integrations with IDs.
	GetBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]models.Integration, error)

	// GetByConfigField returns the Integrations with config fieldName=fieldValue.
	GetByConfigField(ctx context.Context, fieldName string, fieldValue string, DB database.Database) ([]models.Integration, error)

	// GetByNameAndUser returns the Integration named integrationName created by the user with the ID userID in the organization with the ID orgID.
	GetByNameAndUser(ctx context.Context, integrationName string, userID uuid.UUID, orgID string, DB database.Database) (*models.Integration, error)

	// GetByOrg returns the Integrations by the organization with the ID orgID.
	GetByOrg(ctx context.Context, orgID string, DB database.Database) ([]models.Integration, error)

	// GetByServiceAndUser returns the Integrations with the specified service created by the user with the ID userID.
	GetByServiceAndUser(ctx context.Context, service integration.Service, userID uuid.UUID, DB database.Database) ([]models.Integration, error)

	// GetByUser returns the Integrations created by the org where the userID is equal to userID or it is NULL.
	GetByUser(ctx context.Context, orgID string, userID uuid.UUID, DB database.Database) ([]models.Integration, error)

	// ValidateOwnership checks whether the integration is owned by the user if the integration is of type userOnly otherwise, checks whether the integration is owned by the organizaion orgID.
	ValidateOwnership(ctx context.Context, integrationID uuid.UUID, orgID string, userID uuid.UUID, DB database.Database) (bool, error)
}

type integrationWriter interface {
	// Create inserts a new Integration with the specified fields.
	Create(
		ctx context.Context,
		orgID string,
		service integration.Service,
		name string,
		config *utils.Config,
		validated bool,
		DB database.Database,
	) (*models.Integration, error)

	// CreateForUser inserts a new Integration with the specified fields for the given user.
	CreateForUser(
		ctx context.Context,
		orgID string,
		userID uuid.UUID,
		service integration.Service,
		name string,
		config *utils.Config,
		validated bool,
		DB database.Database,
	) (*models.Integration, error)

	// Delete deletes the Integration with ID.
	Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error

	// Update applies changes to the Integration with ID. It returns the updated Integration.
	Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, DB database.Database) (*models.Integration, error)
}
