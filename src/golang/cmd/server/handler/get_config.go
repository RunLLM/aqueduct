package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/config"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/execution_state"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
)

type getConfigArgs struct {
	*aq_context.AqContext
}

type getConfigResponse struct {
	AqPath              string                     `json:"aqPath"`
	RetentionJobPeriod  string                     `json:"retentionJobPeriod"`
	ApiKey              string                     `json:"apiKey"`
	StorageConfigPublic shared.StorageConfigPublic `json:"storageConfig"`
}

type GetConfigHandler struct {
	GetHandler

	ResourceRepo         repos.Resource
	StorageMigrationRepo repos.StorageMigration
	Database             database.Database
}

func (*GetConfigHandler) Name() string {
	return "GetConfig"
}

func (h *GetConfigHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	return &getConfigArgs{
		AqContext: aqContext,
	}, http.StatusOK, nil
}

// TODO(ENG-2725): We should use the database as the source of truth, not the config file.
func (h *GetConfigHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*getConfigArgs)

	storageConfig := config.Storage()
	storageConfigPtr := &storageConfig
	storageConfigPublic, err := storageConfigPtr.ToPublic()
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve storage config.")
	}

	var resourceObj *models.Resource

	// There are a number of fields we need to augment the response with, which aren't directly fetched from
	// the config file. These include resource name, connected-at timestamp, and execution state.
	currStorageMigrationObj, err := h.StorageMigrationRepo.Current(ctx, h.Database)
	if err != nil && !errors.Is(err, database.ErrNoRows()) {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error when fetching current storage resource.")
	}
	if err != nil {
		// If there was no previous storage migration, we must be using the local filesystem.
		resourceObj, err = h.ResourceRepo.GetByNameAndUser(ctx, shared.ArtifactStorageResourceName, args.ID, args.OrgID, h.Database)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error when fetching current storage resource.")
		}
		execState, err := execution_state.ExtractConnectionState(resourceObj)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to fetch status of Filesystem storage resource.")
		}
		storageConfigPublic.ConnectedAt = resourceObj.CreatedAt.Unix()
		storageConfigPublic.ExecState = execState
	} else {
		resourceObj, err = h.ResourceRepo.Get(ctx, currStorageMigrationObj.DestResourceID, h.Database)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error when fetching current storage resource.")
		}
		storageConfigPublic.ConnectedAt = currStorageMigrationObj.ExecState.Timestamps.RegisteredAt.Unix()
		storageConfigPublic.ExecState = &currStorageMigrationObj.ExecState
	}
	storageConfigPublic.ResourceID = resourceObj.ID
	storageConfigPublic.ResourceName = resourceObj.Name

	return getConfigResponse{
		AqPath:              config.AqueductPath(),
		RetentionJobPeriod:  config.RetentionJobPeriod(),
		ApiKey:              config.APIKey(),
		StorageConfigPublic: *storageConfigPublic,
	}, http.StatusOK, nil
}
