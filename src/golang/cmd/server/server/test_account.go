package server

import (
	"context"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/demo"
	"github.com/google/uuid"
)

func CreateTestAccount(
	ctx context.Context,
	s *AqServer,
	apiKey string,
	orgID string,
) (*models.User, error) {
	// Check if test user already exists
	testUser, err := s.UserRepo.GetByAPIKey(ctx, apiKey, s.Database)
	if err != nil && !errors.Is(err, database.ErrNoRows()) {
		return nil, errors.Newf("Unable to check if test account exists: %v", err)
	}

	if errors.Is(err, database.ErrNoRows()) {
		// Create a test user to perform actions from SDK.
		testUser, err = s.UserRepo.Create(
			ctx,
			orgID,
			apiKey,
			s.Database,
		)
		if err != nil {
			return nil, errors.Newf("Unable to create test account: %v", err)
		}
	}

	return testUser, nil
}

func CheckBuiltinIntegration(ctx context.Context, s *AqServer, orgID string) (bool, error) {
	// Check if builtin integration is already connected
	integrations, err := s.IntegrationRepo.GetByOrg(
		context.Background(),
		orgID,
		s.Database,
	)
	if err != nil {
		return false, errors.Newf("Unable to get connected integrations: %v", err)
	}

	for _, integrationObject := range integrations {
		if integrationObject.Name == shared.DemoDbIntegrationName {
			// Builtin integration already connected
			return true, nil
		}
	}

	return false, nil
}

// ConnectBuiltinIntegration adds a builtin integration for the specified
// user's organization. It returns an error, if any.
func ConnectBuiltinIntegration(
	ctx context.Context,
	user *models.User,
	integrationRepo repos.Integration,
	db database.Database,
) error {
	serviceType := shared.Sqlite
	builtinConfig := demo.GetSqliteIntegrationConfig()

	if _, _, err := handler.ConnectIntegration(
		ctx,
		&handler.ConnectIntegrationArgs{
			AqContext: &aq_context.AqContext{
				User:      *user,
				RequestID: uuid.New().String(),
			},
			Name:     shared.DemoDbIntegrationName,
			Service:  serviceType,
			Config:   builtinConfig,
			UserOnly: false,
		},
		integrationRepo,
		db,
	); err != nil {
		return err
	}

	return nil
}
