package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/config"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator/connector"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/scheduler"
	workflow_utils "github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const (
	PollPreviewTableInterval = 500 * time.Millisecond
	PollPreviewTableTimeout  = 60 * time.Second
)

// Route: /integration/{integrationId}/preview
// Method: GET
// Params:
//	`integrationId`: ID of the relational database integration
// Request:
//	Headers:
//		`api-key`: user's API Key
//		`table-name`: name of the table to preview
// Response:
//	Body:
//		serialized `previewTableResponse`, the json serialized table content

type previewTableArgs struct {
	*aq_context.AqContext
	integrationID uuid.UUID
	tableName     string
}

type previewTableResponse struct {
	Data string `json:"data"`
}

type PreviewTableHandler struct {
	GetHandler

	Database   database.Database
	JobManager job.JobManager

	IntegrationRepo repos.Integration
}

func (*PreviewTableHandler) Name() string {
	return "PreviewTable"
}

func (*PreviewTableHandler) Headers() []string {
	return []string{routes.TableNameHeader}
}

func (h *PreviewTableHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	integrationIDStr := chi.URLParam(r, routes.IntegrationIdUrlParam)
	integrationID, err := uuid.Parse(integrationIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed integration ID.")
	}

	tableName := r.Header.Get(routes.TableNameHeader)
	if tableName == "" {
		return nil, http.StatusBadRequest, errors.Wrap(err, "No table name specified.")
	}

	ok, err := h.IntegrationRepo.ValidateOwnership(
		r.Context(),
		integrationID,
		aqContext.OrgID,
		aqContext.ID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during integration ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this integration.")
	}

	return &previewTableArgs{
		AqContext:     aqContext,
		integrationID: integrationID,
		tableName:     tableName,
	}, http.StatusOK, nil
}

func (h *PreviewTableHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*previewTableArgs)

	integrationObject, err := h.IntegrationRepo.Get(
		ctx,
		args.integrationID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to retrieve integration.")
	}

	if !shared.IsRelationalDatabaseIntegration(integrationObject.Service) {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Preview table request is only allowed for relational databases.")
	}

	operatorMetadataPath := fmt.Sprintf("operator-metadata-%s", args.RequestID)
	artifactMetadataPath := fmt.Sprintf("artifact-metadata-%s", args.RequestID)
	artifactContentPath := fmt.Sprintf("artifact-content-%s", args.RequestID)

	defer func() {
		// Delete storage files created for preview table data
		go workflow_utils.CleanupStorageFiles(ctx, args.StorageConfig, []string{operatorMetadataPath, artifactMetadataPath, artifactContentPath})
	}()

	var queryParams connector.ExtractParams
	if integrationObject.Service == shared.MongoDB {
		// This triggers `db.my_table.find({})`
		queryParams = &connector.MongoDBExtractParams{
			Collection:      args.tableName,
			QuerySerialized: "{\"args\": [{}]}",
		}
	} else {
		queryParams = &connector.RelationalDBExtractParams{
			Query: fmt.Sprintf("SELECT * FROM %s;", args.tableName),
		}
	}

	storageConfig := config.Storage()
	vaultObject, err := vault.NewVault(&storageConfig, config.EncryptionKey())
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to initialize vault.")
	}

	jobName, err := scheduler.ScheduleExtract(
		ctx,
		connector.Extract{
			Service:       integrationObject.Service,
			IntegrationId: integrationObject.ID,
			Parameters:    queryParams,
		},
		operatorMetadataPath,
		[]string{}, /* inputParamNames */
		[]string{}, /* inputContentPaths */
		[]string{}, /* inputMetadataPaths */
		artifactContentPath,
		artifactMetadataPath,
		args.StorageConfig,
		h.JobManager,
		vaultObject,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to schedule job to preview table.")
	}

	jobStatus, err := job.PollJob(ctx, jobName, h.JobManager, PollPreviewTableInterval, PollPreviewTableTimeout)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error while waiting for the preview table job to finish.")
	}

	if jobStatus == shared.FailedExecutionStatus {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error while previewing table.")
	}

	var metadata shared.ExecutionState
	if err := workflow_utils.ReadFromStorage(
		ctx,
		args.StorageConfig,
		operatorMetadataPath,
		&metadata,
	); err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve operator metadata from storage.")
	}

	if metadata.Error != nil {
		return nil, http.StatusBadRequest, errors.Newf("Unable to preview table: %v", metadata.Error.Context)
	}

	data, err := storage.NewStorage(args.StorageConfig).Get(
		ctx,
		artifactContentPath,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Failed to retrieve data for the table result.")
	}

	return previewTableResponse{
		Data: string(data),
	}, http.StatusOK, nil
}
