package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aqueducthq/aqueduct/internal/server/utils"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	workflow_utils "github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

const (
	pollDiscoverInterval = 500 * time.Millisecond
	pollDiscoverTimeout  = 60 * time.Second
)

// Route: /{integrationId}/tables
// Method: GET
// Params:
//	`integrationId`: ID of the relational database integration
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		serialized `discoverResponse`, a list of table names

type discoverArgs struct {
	*CommonArgs
	integrationId uuid.UUID
}

type discoverResponse struct {
	TableNames []string `json:"table_names"`
}

type DiscoverHandler struct {
	GetHandler

	Database          database.Database
	IntegrationReader integration.Reader
	StorageConfig     *shared.StorageConfig
	JobManager        job.JobManager
	Vault             vault.Vault
}

func (*DiscoverHandler) Name() string {
	return "Discover"
}

func (h *DiscoverHandler) Prepare(r *http.Request) (interface{}, int, error) {
	common, statusCode, err := ParseCommonArgs(r)
	if err != nil {
		return nil, statusCode, err
	}

	integrationIdStr := chi.URLParam(r, utils.IntegrationIdUrlParam)
	integrationId, err := uuid.Parse(integrationIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed integration ID.")
	}

	ok, err := h.IntegrationReader.ValidateIntegrationOwnership(
		r.Context(),
		integrationId,
		common.OrganizationId,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during integration ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this integration.")
	}

	return &discoverArgs{
		CommonArgs:    common,
		integrationId: integrationId,
	}, http.StatusOK, nil
}

func (h *DiscoverHandler) Perform(
	ctx context.Context,
	interfaceArgs interface{},
) (interface{}, int, error) {
	args := interfaceArgs.(*discoverArgs)

	integrationObject, err := h.IntegrationReader.GetIntegration(
		ctx,
		args.integrationId,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to retrieve integration.")
	}

	if _, ok := integration.GetRelationalDatabaseIntegrations()[integrationObject.Service]; !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "List tables request is only allowed for relational databases.")
	}

	jobMetadataPath := fmt.Sprintf("list-tables-metadata-%s", args.RequestId)
	jobResultPath := fmt.Sprintf("list-tables-result-%s", args.RequestId)

	defer func() {
		// Delete storage files created for list tables job metadata
		go workflow_utils.CleanupStorageFiles(ctx, h.StorageConfig, []string{jobMetadataPath, jobResultPath})
	}()

	config, err := auth.ReadConfigFromSecret(ctx, integrationObject.Id, h.Vault)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to parse integration config.")
	}

	jobName := fmt.Sprintf("discover-operator-%s", uuid.New().String())
	jobSpec := job.NewDiscoverSpec(
		jobName,
		h.StorageConfig,
		jobMetadataPath,
		integrationObject.Service,
		config,
		jobResultPath,
	)

	if err := h.JobManager.Launch(ctx, jobName, jobSpec); err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to launch discover job.")
	}

	jobStatus, err := job.PollJob(ctx, jobName, h.JobManager, pollDiscoverInterval, pollDiscoverTimeout)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error while waiting for discover job to finish.")
	}

	if jobStatus == shared.FailedExecutionStatus {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error while listing tables.")
	}

	var metadata operator_result.Metadata
	if err := workflow_utils.ReadFromStorage(
		ctx,
		h.StorageConfig,
		jobMetadataPath,
		&metadata,
	); err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve operator metadata from storage.")
	}

	if len(metadata.Error) > 0 {
		return nil, http.StatusBadRequest, errors.Newf("Unable to list tables: %v", metadata.Error)
	}

	var tableNames []string
	if err := workflow_utils.ReadFromStorage(
		ctx,
		h.StorageConfig,
		jobResultPath,
		&tableNames,
	); err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve table names from storage.")
	}

	return discoverResponse{
		TableNames: tableNames,
	}, http.StatusOK, nil
}
