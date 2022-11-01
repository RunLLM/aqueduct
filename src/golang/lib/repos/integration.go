package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/google/uuid"
)

// Integration defines all of the database operations that can be performed for a Integration.
type Integration interface {
	integrationReader
	integrationWriter
}

type integrationReader interface {
	// Get returns the Integration with id.
	Get(ctx context.Context, id uuid.UUID, db database.Database) (*models.Integration, error)

	// GetMultiple returns the Integrations with ids.
	GetMultiple(ctx context.Context, ids []uuid.UUID, db database.Database) ([]models.Integration, error)

	// GetByServiceAndUser returns the Integrations with the specified service by the user with the id userId.
	GetByServiceAndUser(ctx context.Context, service Service, userId uuid.UUID, db database.Database) ([]models.Integration, error)

	// GetByOrganization returns the Integrations by the organization with the id organizationId.
	GetByOrganization(ctx context.Context, organizationId string, db database.Database) ([]models.Integration, error)

	// GetByUser returns the Integrations by the user with the id userId in the organization with the id organizationId.
	GetByUser(ctx context.Context, organizationId string, userId uuid.UUID, db database.Database) ([]models.Integration, error)
	
	// GetByNameAndUser returns the Integrations with name integrationName by the user with the id userId in the organization with the id organizationId.
	GetByNameAndUser(ctx context.Context, integrationName string, userId uuid.UUID, organizationId string, db database.Database) ([]models.Integration, error)

	// GetByConfigField returns the Integrations with config fieldName=fieldValue.
	GetByConfigField(ctx context.Context, fieldName string, fieldValue string, db database.Database) ([]models.Integration, error)

	// ValidateOwnership returns the Integrations with config fieldName=fieldValue.
	ValidateOwnership(ctx context.Context, integrationId uuid.UUID, organizationId string, userId uuid.UUID, db database.Database) (bool, error)
}

type integrationWriter interface {
	// Create inserts a new Integration with the specified fields.
	Create(
		ctx context.Context,
		organizationId string,
		service Service,
		name string,
		config *utils.Config,
		validated bool,
		db database.Database,
	) (*models.Integration, error)

	// CreateForUser inserts a new Integration with the specified fields for the given user.
	CreateForUser(
		ctx context.Context,
		organizationId string,
		userId uuid.UUID,
		service Service,
		name string,
		config *utils.Config,
		validated bool,
		db database.Database,
	) (*models.Integration, error)

	// Delete deletes the Integration with id.
	Delete(ctx context.Context, id uuid.UUID, db database.Database) error

	// Update applies changes to the Integration with id. It returns the updated Integration.
	Update(ctx context.Context, id uuid.UUID, changes map[string]interface{}, db database.Database) (*models.Integration, error)
}
