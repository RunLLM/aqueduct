package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/google/uuid"
)

// Integration defines all of the database operations that can be performed for a Integration.
type Integration interface {
	integrationReader
	integrationWriter
}

type integrationReader interface {
	// Get returns the Integration with ID.
	Get(ctx context.Context, ID uuid.UUID, db database.Database) (*models.Integration, error)

	// GetBatch returns the Integrations with IDs.
	GetBatch(ctx context.Context, IDs []uuid.UUID, db database.Database) ([]models.Integration, error)

	// GetByServiceAndUser returns the Integrations with the specified service by the user with the ID userID.
	GetByServiceAndUser(ctx context.Context, service shared.Service, userID uuid.UUID, db database.Database) ([]models.Integration, error)

	// GetByOrganization returns the Integrations by the organization with the ID orgID.
	GetByOrganization(ctx context.Context, orgId string, db database.Database) ([]models.Integration, error)

	// GetByUser returns the Integrations by the user with the ID userID in the organization with the ID orgID.
	GetByUser(ctx context.Context, orgID string, userID uuid.UUID, db database.Database) ([]models.Integration, error)

	// GetByNameAndUser returns the Integrations with name integrationName by the user with the ID userID in the organization with the ID orgID.
	GetByNameAndUser(ctx context.Context, integrationName string, userID uuid.UUID, orgID string, db database.Database) ([]models.Integration, error)

	// GetByConfigField returns the Integrations with config fieldName=fieldValue.
	GetByConfigField(ctx context.Context, fieldName string, fieldValue string, db database.Database) ([]models.Integration, error)

	// ValidateOwnership checks whether the integration is owned by the user.
	ValidateOwnership(ctx context.Context, integrationID uuid.UUID, orgID string, userID uuid.UUID, db database.Database) (bool, error)
}

type integrationWriter interface {
	// Create inserts a new Integration with the specified fields.
	Create(
		ctx context.Context,
		orgID string,
		service shared.Service,
		name string,
		config *utils.Config,
		validated bool,
		db database.Database,
	) (*models.Integration, error)

	// CreateForUser inserts a new Integration with the specified fields for the given user.
	CreateForUser(
		ctx context.Context,
		orgID string,
		userID uuid.UUID,
		service shared.Service,
		name string,
		config *utils.Config,
		validated bool,
		db database.Database,
	) (*models.Integration, error)

	// Delete deletes the Integration with ID.
	Delete(ctx context.Context, ID uuid.UUID, db database.Database) error

	// Update applies changes to the Integration with ID. It returns the updated Integration.
	Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, db database.Database) (*models.Integration, error)
}
