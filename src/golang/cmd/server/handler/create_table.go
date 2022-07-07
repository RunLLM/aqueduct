package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
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
	*aq_context.AqContext
	tableName     string
	integrationId uuid.UUID
	csv           []byte
}

type CreateTableResponse struct{}

var usefulHeaders = map[string]bool{
	"accept-encoding":    true,
	"accept":             true,
	"connection":         true,
	"api-key":            true,
	"sdk-client-version": true,
	"content-length":     true,
	"user-agent":         true,
	"content-type":       true,
	"table-name":         true,
	"transfer-encoding":  true,
}

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

	log.Info("logging headers...")
	for name := range r.Header {
		log.Infof("%s: %s", name, r.Header[name])
	}

	toRemove := []string{}
	// Loop over header names
	for name := range r.Header {
		if _, ok := usefulHeaders[strings.ToLower(name)]; !ok {
			log.Infof("removing header: %s", name)
			toRemove = append(toRemove, name)
		}
	}

	for _, header := range toRemove {
		r.Header.Del(header)
	}

	r.Header.Set("Accept-Encoding", "deflate, gzip")
	r.Header.Set("Transfer-Encoding", "deflate, gzip")

	for name := range r.Header {
		log.Infof("%s: %s", name, r.Header[name])
	}

	log.Info(r.ContentLength)
	r.ContentLength = 1929
	log.Info(r.ContentLength)

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

	integrationObject, err := h.IntegrationReader.GetIntegration(
		ctx,
		args.integrationId,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Cannot get integration.")
	}

	// Save CSV
	contentPath := fmt.Sprintf("create-table-content-%s", args.RequestId)
	csvStorage := storage.NewStorage(h.StorageConfig)
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

	if statusCode, err := CreateTable(ctx, args, contentPath, integrationObject, h.Vault, h.StorageConfig, h.JobManager); err != nil {
		return emptyResp, statusCode, err
	}

	return emptyResp, returnStatus, returnErr
}

// CreateTable adds the CSV as a table in the database. It returns a status code for the request
// and an error, if any.
func CreateTable(ctx context.Context, args *CreateTableArgs, contentPath string, integrationObject *integration.Integration, vaultObject vault.Vault, storageConfig *shared.StorageConfig, jobManager job.JobManager) (int, error) {
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
