package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/config"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	workflow_utils "github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Doesn't currently work for S3 because it's too expensive to list.

// Route: /resource/{resourceID}/objects
// Method: GET
// Params:
//	`resourceID`: ID for `resource` object
// Request:
//	Headers:
//		`api-key`: user's API Key
//
// Response: objects written by workflows at the resource.

// Get objects from the specified resource.
type ListResourceObjectsHandler struct {
	GetHandler

	Database   database.Database
	JobManager job.JobManager

	ResourceRepo repos.Resource
}

type ListResourceObjectsArgs struct {
	*aq_context.AqContext
	resourceID uuid.UUID
}

type ListResourceObjectsResponse struct {
	ObjectNames []string `json:"object_names"`
}

func (*ListResourceObjectsHandler) Name() string {
	return "IntegrationObjects"
}

func (h *ListResourceObjectsHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, errors.Wrap(err, "Unable to parse arguments.")
	}

	resourceIdStr := chi.URLParam(r, routes.IntegrationIdUrlParam)
	resourceId, err := uuid.Parse(resourceIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed resource ID.")
	}

	return &ListResourceObjectsArgs{
		AqContext:  aqContext,
		resourceID: resourceId,
	}, http.StatusOK, nil
}

func (h *ListResourceObjectsHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*ListResourceObjectsArgs)

	resourceObject, err := h.ResourceRepo.Get(
		ctx,
		args.resourceID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to retrieve resource.")
	}

	if !shared.IsRelationalDatabaseResource(resourceObject.Service) {
		return nil, http.StatusBadRequest, errors.New("List objects request is only allowed for relational databases. (Too expensive to list objects for S3)")
	}

	jobMetadataPath := fmt.Sprintf("list-objects-metadata-%s", args.RequestID)
	jobResultPath := fmt.Sprintf("list-objects-result-%s", args.RequestID)

	defer func() {
		// Delete storage files created for list objects job metadata
		go workflow_utils.CleanupStorageFiles(context.Background(), args.StorageConfig, []string{jobMetadataPath, jobResultPath})
	}()

	storageConfig := config.Storage()
	vaultObject, err := vault.NewVault(&storageConfig, config.EncryptionKey())
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to initialize vault.")
	}

	config, err := auth.ReadConfigFromSecret(ctx, resourceObject.ID, vaultObject)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to parse resource config.")
	}

	jobName := fmt.Sprintf("resource-objects-%s", uuid.New().String())
	jobSpec := job.NewDiscoverSpec(
		jobName,
		args.StorageConfig,
		jobMetadataPath,
		resourceObject.Service,
		config,
		jobResultPath,
	)

	if err := h.JobManager.Launch(ctx, jobName, jobSpec); err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to launch resource objects job.")
	}

	jobStatus, err := job.PollJob(ctx, jobName, h.JobManager, pollDiscoverInterval, pollDiscoverTimeout)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error while waiting for resource objects job to finish.")
	}

	if jobStatus == shared.FailedExecutionStatus {
		return nil, http.StatusInternalServerError, errors.New("Unexpected error while listing objects.")
	}

	var metadata shared.ExecutionState
	if err := workflow_utils.ReadFromStorage(
		ctx,
		args.StorageConfig,
		jobMetadataPath,
		&metadata,
	); err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve operator metadata from storage.")
	}

	if metadata.Error != nil {
		return nil, http.StatusBadRequest, errors.Newf("Unable to list objects: %v", metadata.Error.Context)
	}

	var objectNames []string
	if err := workflow_utils.ReadFromStorage(
		ctx,
		args.StorageConfig,
		jobResultPath,
		&objectNames,
	); err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve object names from storage.")
	}

	return ListResourceObjectsResponse{
		ObjectNames: objectNames,
	}, http.StatusOK, nil
}
