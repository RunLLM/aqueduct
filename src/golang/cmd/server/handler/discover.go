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
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	workflow_utils "github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const (
	pollDiscoverInterval = 500 * time.Millisecond
	pollDiscoverTimeout  = 60 * time.Second
)

// Route: /resource/{resourceID}/discover
// Method: GET
// Params:
//	`resourceID`: ID of the relational database resource
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		serialized `discoverResponse`, a list of table names

type discoverArgs struct {
	*aq_context.AqContext
	resourceID uuid.UUID
}

type discoverResponse struct {
	TableNames []string `json:"table_names"`
}

type DiscoverHandler struct {
	GetHandler

	Database   database.Database
	JobManager job.JobManager

	ResourceRepo repos.Resource
	OperatorRepo repos.Operator
}

func (*DiscoverHandler) Name() string {
	return "Discover"
}

func (h *DiscoverHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	resourceIDStr := chi.URLParam(r, routes.ResourceIDUrlParam)
	resourceID, err := uuid.Parse(resourceIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed resource ID.")
	}

	ok, err := h.ResourceRepo.ValidateOwnership(
		r.Context(),
		resourceID,
		aqContext.OrgID,
		aqContext.ID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during resource ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this resource.")
	}

	return &discoverArgs{
		AqContext:  aqContext,
		resourceID: resourceID,
	}, http.StatusOK, nil
}

func (h *DiscoverHandler) Perform(
	ctx context.Context,
	interfaceArgs interface{},
) (interface{}, int, error) {
	args := interfaceArgs.(*discoverArgs)

	resourceObject, err := h.ResourceRepo.Get(
		ctx,
		args.resourceID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to retrieve resource.")
	}

	if !shared.IsRelationalDatabaseResource(resourceObject.Service) {
		return nil, http.StatusBadRequest, errors.Wrap(err, "List tables request is only allowed for relational databases.")
	}

	jobMetadataPath := fmt.Sprintf("list-tables-metadata-%s", args.RequestID)
	jobResultPath := fmt.Sprintf("list-tables-result-%s", args.RequestID)

	defer func() {
		// Delete storage files created for list tables job metadata
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

	jobName := fmt.Sprintf("discover-operator-%s", uuid.New().String())
	jobSpec := job.NewDiscoverSpec(
		jobName,
		args.StorageConfig,
		jobMetadataPath,
		resourceObject.Service,
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
		return nil, http.StatusBadRequest, errors.Newf("Unable to list tables: %v", metadata.Error.Context)
	}

	var tableNames []string
	if err := workflow_utils.ReadFromStorage(
		ctx,
		args.StorageConfig,
		jobResultPath,
		&tableNames,
	); err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve table names from storage.")
	}

	loadOPSpecs, err := h.OperatorRepo.GetLoadOPSpecsByOrg(
		ctx,
		args.OrgID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to get load operators.")
	}

	// All user-created tables.
	userTables := make(map[string]bool, len(loadOPSpecs))
	for _, loadOPSpec := range loadOPSpecs {
		loadSpec, ok := connector.CastToRelationalDBLoadParams(loadOPSpec.Spec.Load().Parameters)
		if !ok {
			return nil, http.StatusInternalServerError, errors.Newf("Cannot load table")
		}
		table := loadSpec.Table
		userTables[table] = true
	}

	baseTables := make([]string, 0, len(tableNames))

	for _, tableName := range tableNames {
		if isUserTable := userTables[tableName]; !isUserTable { // not a user-created table
			baseTables = append(baseTables, tableName)
		}
	}

	return discoverResponse{
		TableNames: baseTables,
	}, http.StatusOK, nil
}
