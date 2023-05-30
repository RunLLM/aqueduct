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

// connectBuiltinResources checks for any missing built-in resources, and connects them if they are missing.
// If the deprecated demo db name still exists in the database, we delete it before connecting the new one.
func connectBuiltinResources(
	ctx context.Context,
	s *AqServer,
	orgID string,
	user *models.User,
	resourceRepo repos.Resource,
	db database.Database,
) error {
	resources, err := s.ResourceRepo.GetByOrg(
		context.Background(),
		orgID,
		s.Database,
	)
	if err != nil {
		return errors.Newf("Unable to get connected resources: %v", err)
	}

	demoConnected := false
	engineConnected := false
	filesystemConnected := false
	for _, resourceObject := range resources {
		if resourceObject.Name == shared.DeprecatedDemoDBResourceName && resourceObject.Service == shared.Sqlite {
			if err := s.ResourceRepo.Delete(
				ctx,
				resourceObject.ID,
				s.Database,
			); err != nil {
				return errors.Newf("Unable to delete deprecated demo resource: %v", err)
			}
			continue
		} else if resourceObject.Name == shared.DemoDbName {
			demoConnected = true
		} else if resourceObject.Name == shared.AqueductComputeName {
			engineConnected = true
		} else if resourceObject.Name == shared.ArtifactStorageResourceName {
			filesystemConnected = true
		}

		if demoConnected && engineConnected && filesystemConnected {
			// Builtin resources already connected
			return nil
		}
	}

	if !demoConnected {
		err = connectBuiltinDemoDBResource(ctx, user, resourceRepo, db)
		if err != nil {
			return err
		}
	}

	if !engineConnected {
		err = connectBuiltinComputeResource(ctx, user, resourceRepo, db)
		if err != nil {
			return err
		}
	}
	if !filesystemConnected {
		err = connectBuiltinArtifactStorageResource(ctx, user, resourceRepo, db)
		if err != nil {
			return err
		}
	}
	return nil
}

// connectBuiltinDemoDBResource adds the builtin demo data resources for the specified
// user's organization. It returns an error, if any.
func connectBuiltinDemoDBResource(
	ctx context.Context,
	user *models.User,
	resourceRepo repos.Resource,
	db database.Database,
) error {
	builtinConfig := demo.GetSqliteResourceConfig()
	if _, _, err := handler.ConnectResource(
		ctx,
		nil, // Not registering an AWS resource.
		&handler.ConnectResourceArgs{
			AqContext: &aq_context.AqContext{
				User:      *user,
				RequestID: uuid.New().String(),
			},
			Name:     shared.DemoDbName,
			Service:  shared.Sqlite,
			Config:   builtinConfig,
			UserOnly: false,
		},
		resourceRepo,
		db,
	); err != nil {
		return err
	}

	return nil
}

func connectBuiltinComputeResource(
	ctx context.Context,
	user *models.User,
	resourceRepo repos.Resource,
	db database.Database,
) error {
	if _, _, err := handler.ConnectResource(
		ctx,
		nil, // Not registering an AWS resource.
		&handler.ConnectResourceArgs{
			AqContext: &aq_context.AqContext{
				User:      *user,
				RequestID: uuid.New().String(),
			},
			Name:     shared.AqueductComputeName,
			Service:  shared.Aqueduct,
			Config:   auth.NewStaticConfig(map[string]string{}),
			UserOnly: false,
		},
		resourceRepo,
		db,
	); err != nil {
		return err
	}
	return nil
}

func connectBuiltinArtifactStorageResource(
	ctx context.Context,
	user *models.User,
	resourceRepo repos.Resource,
	db database.Database,
) error {
	// TODO(ENG-2941): This is currently duplicated in src/golang/config/config.go.
	defaultStoragePath := path.Join(os.Getenv("HOME"), ".aqueduct", "server", "storage")

	if _, _, err := handler.ConnectResource(
		ctx,
		nil, // Not registering an AWS resource.
		&handler.ConnectResourceArgs{
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
		resourceRepo,
		db,
	); err != nil {
		return err
	}
	return nil
}
