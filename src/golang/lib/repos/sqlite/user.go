package sqlite

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/google/uuid"
)

type userRepo struct {
	userReader
	userWriter
}

type userReader struct{}

type userWriter struct{}

func NewUserRepo() repos.User {
	return &userRepo{
		userReader: userReader{},
		userWriter: userWriter{},
	}
}

func (*userReader) GetByAPIKey(ctx context.Context, apiKey string, DB database.Database) (*models.User, error) {
	return nil, nil
}

func (*userReader) GetOrgAdmin(ctx context.Context, orgID string, DB database.Database) (*models.User, error) {
	return nil, nil
}

func (*userWriter) Create(
	ctx context.Context,
	email string,
	orgID string,
	role string,
	apiKey string,
	auth0ID string,
) (*models.User, error) {
	return nil, nil
}

func (*userWriter) ResetAPIKey(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.User, error) {
	return nil, nil
}
