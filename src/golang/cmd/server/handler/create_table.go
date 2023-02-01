package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/collections/operator/connector"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Route: /integration/{integrationId}/create
// Method: POST
// Params:
//	`integrationId`: ID for `integration` object
//	** ONLY SUPPORTS CREATING TABLES FOR THE DEMO DB **
// Request:
//	Headers:
//		`table-name`: name of table to create
//		`api-key`: user's API Key
// Body:
//		the CSV file to upload to the integration.

const (
	pollCreateInterval = 500 * time.Millisecond
	pollCreateTimeout  = 2 * time.Minute
)

// Creates a table in the specified integration.
type CreateTableHandler struct {
	PostHandler

	Database   database.Database
	JobManager job.JobManager

	IntegrationRepo repos.Integration
}

type CreateTableArgs struct {
	*aq_context.AqContext
	tableName     string
	integrationId uuid.UUID
	csv           []byte
}

type CreateTableResponse struct{}

func (*CreateTableHandler) Name() string {
	return "CreateTable"
}

func (h *CreateTableHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, errors.Wrap(err, "Unable to parse arguments.")
	}
	tableName := r.Header.Get(routes.TableNameHeader)
	integrationIdStr := chi.URLParam(r, routes.IntegrationIdUrlParam)
	integrationId, err := uuid.Parse(integrationIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed integration ID.")
	}

	csv, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to read CSV content.")
	}

	return &CreateTableArgs{
		AqContext:     aqContext,
		tableName:     tableName,
		integrationId: integrationId,
		csv:           csv,
	}, http.StatusOK, nil
}

func (h *CreateTableHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*CreateTableArgs)

	integrationObject, err := h.IntegrationRepo.Get(
		ctx,
		args.integrationId,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Cannot get integration.")
	}

	// Save CSV
	contentPath := fmt.Sprintf("create-table-content-%s", args.RequestID)
	csvStorage := storage.NewStorage(args.StorageConfig)
	if err := csvStorage.Put(ctx, contentPath, args.csv); err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Cannot save CSV.")
	}

	var returnErr error = nil
	returnStatus := http.StatusOK

	defer func() {
		if deleteErr := csvStorage.Delete(ctx, contentPath); deleteErr != nil {
			returnErr = errors.Wrap(deleteErr, "Error deleting CSV from temporary storage.")
			returnStatus = http.StatusInternalServerError
		}
	}()

	emptyResp := CreateTableResponse{}

	storageConfig := config.Storage()
	vaultObject, err := vault.NewVault(&storageConfig, config.EncryptionKey())
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to initialize vault.")
	}

	if statusCode, err := CreateTable(ctx, args, contentPath, integrationObject, vaultObject, args.StorageConfig, h.JobManager); err != nil {
		return emptyResp, statusCode, err
	}

	return emptyResp, returnStatus, returnErr
}

// CreateTable adds the CSV as a table in the database. It returns a status code for the request
// and an error, if any.
func CreateTable(
	ctx context.Context,
	args *CreateTableArgs,
	contentPath string,
	integrationObject *models.Integration,
	vaultObject vault.Vault,
	storageConfig *shared.StorageConfig,
	jobManager job.JobManager,
) (int, error) {
	// Schedule load table job
	jobMetadataPath := fmt.Sprintf("create-table-%s", args.RequestID)

	jobName := fmt.Sprintf("create-table-operator-%s", uuid.New().String())

	config, err := auth.ReadConfigFromSecret(ctx, integrationObject.ID, vaultObject)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to launch create table job.")
	}

	// Assuming service supports GenericRelationalDBLoadParams
	loadParameters := &connector.GenericRelationalDBLoadParams{
		RelationalDBLoadParams: connector.RelationalDBLoadParams{
			Table:      args.tableName,
			UpdateMode: "fail",
		},
	}

	jobSpec := job.NewLoadTableSpec(
		jobName,
		contentPath,
		storageConfig,
		jobMetadataPath,
		integrationObject.Service,
		config,
		loadParameters,
		"",
		"",
	)
	if err := jobManager.Launch(ctx, jobName, jobSpec); err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to launch create table job.")
	}

	jobStatus, err := job.PollJob(ctx, jobName, jobManager, pollCreateInterval, pollCreateTimeout)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to create table.")
	}

	if jobStatus == shared.SucceededExecutionStatus {
		// Table creation was successful
		return http.StatusOK, nil
	}

	// Table creation failed, so we need to fetch the error message from storage
	var execState shared.ExecutionState
	if err := utils.ReadFromStorage(
		ctx,
		storageConfig,
		jobMetadataPath,
		&execState,
	); err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to create table.")
	}

	if execState.Error != nil {
		return http.StatusBadRequest, errors.Newf(
			"Unable to create table.\n%s\n%s",
			execState.Error.Tip,
			execState.Error.Context,
		)
	}

	return http.StatusInternalServerError, errors.New(
		"Unable to create table, we couldn't obtain more context at this point.",
	)
}
