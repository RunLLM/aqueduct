package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/google/uuid"
)

// User defines all of the database operations that can be performed for a User.
type User interface {
	userReader
	userWriter
}

type userReader interface {
	// GetByAPIKey returns the User with the API key apiKey.
	GetByAPIKey(ctx context.Context, apiKey string, DB database.Database) (*models.User, error)

	// GetOrgAdmin returns the admin User for the organization orgID.
	GetOrgAdmin(ctx context.Context, orgID string, DB database.Database) (*models.User, error)
}

type userWriter interface {
	// Creates inserts a new User with the specified fields.
	Create(
		ctx context.Context,
		email string,
		orgID string,
		role string,
		apiKey string,
		auth0ID string,
	) (*models.User, error)

	// ResetAPIKey resets the API key for the User with ID.
	ResetAPIKey(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.User, error)
}
