package handler

import (
	"context"
	"net/http"
	"path"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/config"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/storage_migration"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Route: /config/storage/{resourceID}
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

	ArtifactRepo         repos.Artifact
	ArtifactResultRepo   repos.ArtifactResult
	DAGRepo              repos.DAG
	ResourceRepo         repos.Resource
	OperatorRepo         repos.Operator
	StorageMigrationRepo repos.StorageMigration

	PauseServerFn   func()
	RestartServerFn func()
}

type configureStorageArgs struct {
	*aq_context.AqContext
	// This is the ID of the resource to use as the new storage layer.
	// It should only be set if configureLocalStorage is false.
	storageResourceID     uuid.UUID
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

	resourceIDStr := chi.URLParam(r, routes.IntegrationIdUrlParam)

	if resourceIDStr != "local" {
		return nil, http.StatusBadRequest, errors.Wrap(err, "We currently only support changing the storage layer to the local filesystem from this route.")
	}

	return &configureStorageArgs{
		AqContext: aqContext,
		// TODO ENG-2574: Add support for switching to non-local storage
		storageResourceID:     uuid.Nil,
		configureLocalStorage: true,
	}, http.StatusOK, nil
}

func (h *ConfigureStorageHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*configureStorageArgs)

	// TODO ENG-2574: Remove this assumption
	if !args.configureLocalStorage {
		return nil, http.StatusBadRequest, errors.New("We currently only support changing the storage layer to the local filesystem from this route.")
	}

	currentStorageConfig := config.Storage()

	if currentStorageConfig.Type == shared.FileStorageType {
		return nil, http.StatusBadRequest, errors.New("The storage layer is already set to the local filesystem.")
	}

	newStorageConfig := shared.StorageConfig{
		Type: shared.FileStorageType,
		FileConfig: &shared.FileConfig{
			Directory: path.Join(config.AqueductPath(), "storage"),
		},
	}

	err := storage_migration.Perform(
		ctx,
		args.OrgID,
		nil, /* destResourceObj */
		&newStorageConfig,
		h.PauseServerFn,
		h.RestartServerFn,
		h.ArtifactRepo,
		h.ArtifactResultRepo,
		h.DAGRepo,
		h.ResourceRepo,
		h.OperatorRepo,
		h.StorageMigrationRepo,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to migrate storage layer.")
	}
	return nil, http.StatusOK, nil
}
