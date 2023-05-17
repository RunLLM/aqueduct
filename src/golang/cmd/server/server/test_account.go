package server

import (
	"context"
	"os"
	"path"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
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

// CheckBuiltinIntegrations returns whether the builtin demo and compute integrations already exist.
// If we notice that the deprecated demo integration exists, we delete it. We expect the caller to add
// the appropriate demo integration with `connectBuiltinDemoDBIntegration()` next.
func CheckBuiltinIntegrations(ctx context.Context, s *AqServer, orgID string) (bool, bool, error) {
	integrations, err := s.IntegrationRepo.GetByOrg(
		context.Background(),
		orgID,
		s.Database,
	)
	if err != nil {
		return false, false, errors.Newf("Unable to get connected integrations: %v", err)
	}

	demoConnected := false
	engineConnected := false
	for _, integrationObject := range integrations {
		if integrationObject.Name == shared.DeprecatedDemoDBResourceName && integrationObject.Service == shared.Sqlite {
			if err := s.IntegrationRepo.Delete(
				ctx,
				integrationObject.ID,
				s.Database,
			); err != nil {
				return false, false, errors.Newf("Unable to delete deprecated demo integration: %v", err)
			}
			continue
		} else if integrationObject.Name == shared.DemoDbName {
			demoConnected = true
		} else if integrationObject.Name == shared.AqueductComputeName {
			engineConnected = true
		}

		if demoConnected && engineConnected {
			// Builtin integrations already connected
			return true, true, nil
		}
	}

	return demoConnected, engineConnected, nil
}

// connectBuiltinResources checks for any missing built-in integrations, and connects them if they are missing.
// If the deprecated demo db name still exists in the database, we delete it before connecting the new one.
func connectBuiltinResources(
	ctx context.Context,
	s *AqServer,
	orgID string,
	user *models.User,
	integrationRepo repos.Integration,
	db database.Database,
) error {
	integrations, err := s.IntegrationRepo.GetByOrg(
		context.Background(),
		orgID,
		s.Database,
	)
	if err != nil {
		return errors.Newf("Unable to get connected integrations: %v", err)
	}

	demoConnected := false
	engineConnected := false
	filesystemConnected := false
	for _, integrationObject := range integrations {
		if integrationObject.Name == shared.DeprecatedDemoDBResourceName && integrationObject.Service == shared.Sqlite {
			if err := s.IntegrationRepo.Delete(
				ctx,
				integrationObject.ID,
				s.Database,
			); err != nil {
				return errors.Newf("Unable to delete deprecated demo integration: %v", err)
			}
			continue
		} else if integrationObject.Name == shared.DemoDbName {
			demoConnected = true
		} else if integrationObject.Name == shared.AqueductComputeName {
			engineConnected = true
		} else if integrationObject.Name == shared.ArtifactStorageResourceName {
			filesystemConnected = true
		}

		if demoConnected && engineConnected && filesystemConnected {
			// Builtin resources already connected
			return nil
		}
	}

	if !demoConnected {
		err = connectBuiltinDemoDBIntegration(ctx, user, integrationRepo, db)
		if err != nil {
			return err
		}
	}

	if !engineConnected {
		err = connectBuiltinComputeIntegration(ctx, user, integrationRepo, db)
		if err != nil {
			return err
		}
	}
	if !filesystemConnected {
		err = connectBuiltinArtifactStorageIntegration(ctx, user, integrationRepo, db)
		if err != nil {
			return err
		}
	}
	return nil
}

// connectBuiltinDemoDBIntegration adds the builtin demo data integrations for the specified
// user's organization. It returns an error, if any.
func connectBuiltinDemoDBIntegration(
	ctx context.Context,
	user *models.User,
	integrationRepo repos.Integration,
	db database.Database,
) error {
	builtinConfig := demo.GetSqliteIntegrationConfig()
	if _, _, err := handler.ConnectIntegration(
		ctx,
		nil, // Not registering an AWS integration.
		&handler.ConnectIntegrationArgs{
			AqContext: &aq_context.AqContext{
				User:      *user,
				RequestID: uuid.New().String(),
			},
			Name:     shared.DemoDbName,
			Service:  shared.Sqlite,
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

func connectBuiltinComputeIntegration(
	ctx context.Context,
	user *models.User,
	integrationRepo repos.Integration,
	db database.Database,
) error {
	if _, _, err := handler.ConnectIntegration(
		ctx,
		nil, // Not registering an AWS integration.
		&handler.ConnectIntegrationArgs{
			AqContext: &aq_context.AqContext{
				User:      *user,
				RequestID: uuid.New().String(),
			},
			Name:     shared.AqueductComputeName,
			Service:  shared.Aqueduct,
			Config:   auth.NewStaticConfig(map[string]string{}),
			UserOnly: false,
		},
		integrationRepo,
		db,
	); err != nil {
		return err
	}
	return nil
}

func connectBuiltinArtifactStorageIntegration(
	ctx context.Context,
	user *models.User,
	integrationRepo repos.Integration,
	db database.Database,
) error {
	// TODO(ENG-2941): This is currently duplicated in src/golang/config/config.go.
	defaultStoragePath := path.Join(os.Getenv("HOME"), ".aqueduct", "server", "storage")

	if _, _, err := handler.ConnectIntegration(
		ctx,
		nil, // Not registering an AWS integration.
		&handler.ConnectIntegrationArgs{
			AqContext: &aq_context.AqContext{
				User:      *user,
				RequestID: uuid.New().String(),
			},
			Name:    shared.ArtifactStorageResourceName,
			Service: shared.Filesystem,
			Config: auth.NewStaticConfig(map[string]string{
				"location": defaultStoragePath,
			}),
			UserOnly: false,
		},
		integrationRepo,
		db,
	); err != nil {
		return err
	}
	return nil
}
