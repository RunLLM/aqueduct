package handler

import (
	"context"
	"net/http"
	"path"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/config"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Route: /config/storage/{integrationID}
// Method: POST
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//
// Response: none
type ConfigureStorageHandler struct {
	PostHandler

	Database database.Database
	Engine   engine.Engine

	ArtifactRepo repos.Artifact
	DAGRepo      repos.DAG
	DAGEdgeRepo  repos.DAGEdge
	OperatorRepo repos.Operator
	WorkflowRepo repos.Workflow
}

type configureStorageArgs struct {
	*aq_context.AqContext
	// This is the ID of the integration to use as the new storage layer.
	// It should only be set if configureLocalStorage is false.
	storageIntegrationID  uuid.UUID
	configureLocalStorage bool
}

func (*ConfigureStorageHandler) Name() string {
	return "ConfigureStorage"
}

func (h *ConfigureStorageHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, errors.Wrap(err, "Unable to configure storage layer.")
	}

	integrationIDStr := chi.URLParam(r, routes.IntegrationIdUrlParam)

	if integrationIDStr != "local" {
		return nil, http.StatusBadRequest, errors.Wrap(err, "We currently only support changing the storage layer to the local filesystem from this route.")
	}

	return &configureStorageArgs{
		AqContext: aqContext,
		// TODO: Add support for switching to non-local storage
		configureLocalStorage: true,
	}, http.StatusOK, nil
}

func (h *ConfigureStorageHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*configureStorageArgs)

	// TODO: Remove this assumption
	if !args.configureLocalStorage {
		return nil, http.StatusBadRequest, errors.New("We currently only support changing the storage layer to the local filesystem from this route.")
	}

	storageConfig := shared.StorageConfig{
		Type: shared.FileStorageType,
		FileConfig: &shared.FileConfig{
			Directory: path.Join(config.AqueductPath(), "storage"),
		},
	}

	currentStorageConfig := config.Storage()

	utils.MigrateStorageAndVault(
		ctx,
		&currentStorageConfig,
	)

	return nil, -1, nil
}
