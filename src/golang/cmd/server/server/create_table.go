package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	gateway_utils "github.com/aqueducthq/aqueduct/internal/server/utils"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi"
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

	Database          database.Database
	IntegrationReader integration.Reader
	StorageConfig     *shared.StorageConfig
	JobManager        job.JobManager
	Vault             vault.Vault
}

type CreateTableArgs struct {
	*CommonArgs
	tableName     string
	integrationId uuid.UUID
	csv           string
}

type CreateTableResponse struct{}

func (*CreateTableHandler) Name() string {
	return "CreateTable"
}

func (h *CreateTableHandler) Prepare(r *http.Request) (interface{}, int, error) {
	common, statusCode, err := ParseCommonArgs(r)
	if err != nil {
		return nil, statusCode, errors.Wrap(err, "Unable to parse arguments.")
	}

	tableName := chi.URLParam(r, gateway_utils.IntegrationIdUrlParam)
	integrationIdStr := chi.URLParam(r, gateway_utils.IntegrationIdUrlParam)
	integrationId, err := uuid.Parse(integrationIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed integration ID.")
	}

	csvBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to read CSV content.")
	}
	csv := string(csvBytes)

	return &CreateTableArgs{
		CommonArgs:    common,
		tableName:     tableName,
		integrationId: integrationId,
		csv:           csv,
	}, http.StatusOK, nil
}

func (h *CreateTableHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*CreateTableArgs)

	integrationObject, err := h.IntegrationReader.GetIntegration(
		ctx,
		args.integrationId,
		h.Database,
	)

	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Cannot get integration.")
	}

	emptyResp := CreateTableResponse{}

	if statusCode, err := CreateTable(ctx, args, integrationObject, h.Vault, h.StorageConfig, h.JobManager); err != nil {
		return emptyResp, statusCode, err
	}

	return emptyResp, http.StatusOK, nil
}

// CreateTable adds the CSV as a table in the database. It returns a status code for the request
// and an error, if any.
func CreateTable(ctx context.Context, args *CreateTableArgs, integrationObject *integration.Integration, vaultObject vault.Vault, storageConfig *shared.StorageConfig, jobManager job.JobManager) (int, error) {
	// Schedule load table job
	jobMetadataPath := fmt.Sprintf("create-table-%s", args.RequestId)

	jobName := fmt.Sprintf("create-table-operator-%s", uuid.New().String())

	config, err := auth.ReadConfigFromSecret(ctx, integrationObject.Id, vaultObject)
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
		args.csv,
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
	var metadata operator_result.Metadata
	if err := utils.ReadFromStorage(
		ctx,
		storageConfig,
		jobMetadataPath,
		&metadata,
	); err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to create table.")
	}

	return http.StatusBadRequest, errors.Newf("Unable to create table: %v", metadata.Error)
}
