package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/config"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	exec_env "github.com/aqueducthq/aqueduct/lib/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/functional/slices"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator/connector"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	workflow_utils "github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	pollDeleteSavedObjectsInterval = 500 * time.Millisecond
	pollDeleteSavedObjectsTimeout  = 2 * time.Minute
)

type SavedObjectResult struct {
	Name   string                `json:"name"`
	Result shared.ExecutionState `json:"exec_state"`
}

// Route: /workflow/{workflowId}/delete
// Method: POST
// Params: workflowId
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//	Body:
//		json-serialized `deleteWorkflowInput` object.
//
// Response: json-serialized `deleteWorkflowResponse` object.
//
// The `DeleteWorkflowHandler` does a best effort at deleting a workflow and its dependencies, such as
// k8s resources, Postgres state, and output objects in the user's data warehouse.
type deleteWorkflowArgs struct {
	*aq_context.AqContext
	WorkflowID     uuid.UUID
	ExternalDelete map[string][]string
	Force          bool
}

type deleteWorkflowInput struct {
	// This is a map from integration_id to the serialized load spec we want to delete.
	ExternalDeleteLoadParams map[string][]string `json:"external_delete"`
	// `Force` serve as a safe-guard for client to confirm the deletion.
	// If `Force` is true, all objects specified in `ExternalDelete` field
	// will be removed. Otherwise, we will not delete the objects.
	Force bool `json:"force"`
}

type deleteWorkflowResponse struct {
	// This is a map from integration_id to a list of `SavedObjectResult`
	// implying if each object is successfully deleted.
	SavedObjectDeletionResults map[string][]SavedObjectResult `json:"saved_object_deletion_results"`
}

type DeleteWorkflowHandler struct {
	PostHandler

	Database   database.Database
	Engine     engine.Engine
	JobManager job.JobManager

	ExecutionEnvironmentRepo repos.ExecutionEnvironment
	IntegrationRepo          repos.Integration
	OperatorRepo             repos.Operator
	WorkflowRepo             repos.Workflow
	DagRepo                  repos.DAG
	ArtifactResultRepo       repos.ArtifactResult
}

func (*DeleteWorkflowHandler) Name() string {
	return "DeleteWorkflow"
}

func (h *DeleteWorkflowHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statuscode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statuscode, err
	}

	workflowIDStr := chi.URLParam(r, routes.WorkflowIdUrlParam)
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed workflow ID.")
	}

	ok, err := h.WorkflowRepo.ValidateOrg(
		r.Context(),
		workflowID,
		aqContext.OrgID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during workflow ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.New("The organization does not own this workflow.")
	}

	var input deleteWorkflowInput
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to parse JSON input.")
	}

	// Convert the supplied load params into object identifiers (eg. object names for relational databases)
	externalDelete := make(map[string][]string, len(input.ExternalDeleteLoadParams))
	for integrationName, loadSpecStrList := range input.ExternalDeleteLoadParams {
		for _, loadSpecStr := range loadSpecStrList {
			var loadSpec connector.Load
			err = json.Unmarshal([]byte(loadSpecStr), &loadSpec)
			if err != nil {
				return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to parse request.")
			}

			if relationalLoadParams, ok := connector.CastToRelationalDBLoadParams(loadSpec.Parameters); ok {
				externalDelete[integrationName] = append(externalDelete[integrationName], relationalLoadParams.Table)
			} else if s3LoadParams, ok := loadSpec.Parameters.(*connector.S3LoadParams); ok {
				externalDelete[integrationName] = append(externalDelete[integrationName], s3LoadParams.Filepath)
			} else {
				return nil, http.StatusBadRequest, errors.Newf("Unsupported integration type for deleting objects: %s", integrationName)
			}
		}
	}

	return &deleteWorkflowArgs{
		AqContext:      aqContext,
		WorkflowID:     workflowID,
		ExternalDelete: externalDelete,
		Force:          input.Force,
	}, http.StatusOK, nil
}

func (h *DeleteWorkflowHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*deleteWorkflowArgs)

	resp := deleteWorkflowResponse{}
	resp.SavedObjectDeletionResults = map[string][]SavedObjectResult{}

	nameToID := make(map[string]uuid.UUID, len(args.ExternalDelete))
	for integrationName := range args.ExternalDelete {
		integrationObject, err := h.IntegrationRepo.GetByNameAndUser(
			ctx,
			integrationName,
			args.AqContext.ID,
			args.AqContext.OrgID,
			h.Database,
		)
		if err != nil {
			return resp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while getting integration.")
		}
		nameToID[integrationName] = integrationObject.ID
	}

	// Check objects in list are valid
	objCount := 0

	// This only ever needs to be loaded in the case we are potentially dealing with parameterized saves.
	// These save operators have any parameterized values filled in.
	var saveOpsByIntegrationName map[string][]views.LoadOperator

	for integrationName, savedObjectList := range args.ExternalDelete {
		for _, name := range savedObjectList {
			touchedOperators, err := h.OperatorRepo.GetLoadOPsByWorkflowAndIntegration(
				ctx,
				args.WorkflowID,
				nameToID[integrationName],
				name,
				h.Database,
			)
			if err != nil {
				return resp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while validating objects.")
			}

			// No operator had touched the object at the specified integration. However, this does not mean that the object is
			// invalid. We need to consider the case where the object was part of a parameterized saved.
			var touchedOperatorLoadParams []connector.LoadParams
			if len(touchedOperators) == 0 {
				// Fetch and cache this map once for performance.
				if saveOpsByIntegrationName == nil {
					saveOpsByIntegrationName = make(map[string][]views.LoadOperator)
					saveOpsList, err := GetDistinctSaveOpsByWorkflow(
						ctx,
						args.WorkflowID,
						h.OperatorRepo,
						h.DagRepo,
						h.ArtifactResultRepo,
						h.Database,
					)
					if err != nil {
						return resp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while validating objects.")
					}

					for _, saveOp := range saveOpsList {
						saveOpsByIntegrationName[saveOp.IntegrationName] = append(saveOpsByIntegrationName[saveOp.IntegrationName], saveOpsList...)
					}
				}

				// Check for existence in the parameter-expanded list.
				for _, saveOp := range saveOpsByIntegrationName[integrationName] {
					relationalLoad, ok := connector.CastToRelationalDBLoadParams(saveOp.Spec.Parameters)
					if !ok {
						return resp, http.StatusBadRequest, errors.New("Unexpected error occurred when validating objects.")
					}

					if relationalLoad.Table == name {
						touchedOperatorLoadParams = append(touchedOperatorLoadParams, saveOp.Spec.Parameters)
					}
				}
				// If, after this more thorough check, there still aren't any valid operators, then we give up.
				if len(touchedOperatorLoadParams) == 0 {
					return resp, http.StatusBadRequest, errors.New("Object list not valid. Make sure all objects are touched by the workflow.")
				}
			} else {
				touchedOperatorLoadParams = slices.Map(
					touchedOperators,
					func(operator models.Operator) connector.LoadParams {
						return operator.Spec.Load().Parameters
					},
				)
			}

			if !args.Force {
				// Check none have UpdateMode=append.
				for _, touchedLoadParams := range touchedOperatorLoadParams {
					relationalLoad, ok := connector.CastToRelationalDBLoadParams(touchedLoadParams)
					// Check not updating anything in the integration.
					if ok {
						if relationalLoad.UpdateMode == "append" {
							return resp, http.StatusBadRequest, errors.New("Some objects(s) in list were updated in append mode. If you are sure you want to delete everything, set `force=True`.")
						}
					} else if googleSheets, ok := touchedLoadParams.(*connector.GoogleSheetsLoadParams); ok {
						if googleSheets.SaveMode == "NEWSHEET" {
							return resp, http.StatusBadRequest, errors.New("Some objects(s) in list were updated in append mode. If you are sure you want to delete everything, set `force=True`.")
						}
					}
				}
			}
			objCount += 1
		}
	}

	storageConfig := config.Storage()
	vaultObject, err := vault.NewVault(&storageConfig, config.EncryptionKey())
	if err != nil {
		return resp, http.StatusInternalServerError, errors.Wrap(err, "Unable to initialize vault.")
	}

	// Delete associated objects.
	if objCount > 0 {
		savedObjectDeletionResults, httpResponse, err := DeleteSavedObject(
			ctx,
			args,
			nameToID,
			vaultObject,
			args.StorageConfig,
			h.JobManager,
			h.Database,
			h.IntegrationRepo,
		)
		if httpResponse != http.StatusOK {
			return resp, httpResponse, err
		}
		resp.SavedObjectDeletionResults = savedObjectDeletionResults
	}

	err = h.Engine.DeleteWorkflow(ctx, args.WorkflowID)
	if err != nil {
		return resp, http.StatusInternalServerError, errors.Wrap(err, "Unable to delete workflow.")
	}

	// Check unused conda environments and garbage collect them.
	go func() {
		db, err := database.NewDatabase(h.Database.Config())
		if err != nil {
			log.Errorf("Error creating DB in go routine: %v", err)
			return
		}

		err = exec_env.CleanupUnusedEnvironments(
			context.Background(),
			h.OperatorRepo,
			db,
		)
		if err != nil {
			log.Errorf("%v", err)
		}
	}()

	return resp, http.StatusOK, nil
}

func DeleteSavedObject(
	ctx context.Context,
	args *deleteWorkflowArgs,
	integrationNameToID map[string]uuid.UUID,
	vaultObject vault.Vault,
	storageConfig *shared.StorageConfig,
	jobManager job.JobManager,
	DB database.Database,
	integrationRepo repos.Integration,
) (map[string][]SavedObjectResult, int, error) {
	emptySavedObjectDeletionResults := make(map[string][]SavedObjectResult, 0)

	// Schedule delete written objects job
	jobMetadataPath := fmt.Sprintf("delete-saved-objects-%s", args.RequestID)

	jobName := fmt.Sprintf("delete-saved-objects-%s", uuid.New().String())
	contentPath := fmt.Sprintf("delete-saved-objects-content-%s", args.RequestID)

	defer func() {
		// Delete storage files created for delete saved objects job metadata
		go workflow_utils.CleanupStorageFiles(context.Background(), storageConfig, []string{jobMetadataPath, contentPath})
	}()

	integrationConfigs := make(map[string]auth.Config, len(integrationNameToID))
	integrationNames := make(map[string]shared.Service, len(integrationNameToID))
	for integrationName := range args.ExternalDelete {
		integrationId := integrationNameToID[integrationName]
		config, err := auth.ReadConfigFromSecret(ctx, integrationId, vaultObject)
		if err != nil {
			return emptySavedObjectDeletionResults, http.StatusInternalServerError, errors.Wrap(err, "Unable to get integration configs.")
		}
		integrationConfigs[integrationName] = config
		integrationObjects, err := integrationRepo.GetBatch(ctx, []uuid.UUID{integrationId}, DB)
		if err != nil {
			return emptySavedObjectDeletionResults, http.StatusInternalServerError, errors.Wrap(err, "Unable to get integration configs.")
		}
		if len(integrationObjects) != 1 {
			return emptySavedObjectDeletionResults, http.StatusInternalServerError, errors.New("Unable to get integration configs.")
		}
		integrationNames[integrationName] = integrationObjects[0].Service
	}

	jobSpec := job.NewDeleteSavedObjectsSpec(
		jobName,
		storageConfig,
		jobMetadataPath,
		integrationNames,
		integrationConfigs,
		args.ExternalDelete,
		contentPath,
	)
	if err := jobManager.Launch(ctx, jobName, jobSpec); err != nil {
		return emptySavedObjectDeletionResults, http.StatusInternalServerError, errors.Wrap(err, "Unable to launch delete saved objects job.")
	}

	jobStatus, err := job.PollJob(ctx, jobName, jobManager, pollDeleteSavedObjectsInterval, pollDeleteSavedObjectsTimeout)
	if err != nil {
		return emptySavedObjectDeletionResults, http.StatusInternalServerError, errors.Wrap(err, "Unable to delete saved objects.")
	}

	if jobStatus == shared.SucceededExecutionStatus {
		// Object deletion attempts were successful
		jobSavedObjectDeletionResults := map[string][]SavedObjectResult{}

		if err := workflow_utils.ReadFromStorage(
			ctx,
			storageConfig,
			contentPath,
			&jobSavedObjectDeletionResults,
		); err != nil {
			return emptySavedObjectDeletionResults, http.StatusInternalServerError, errors.Wrap(err, "Unable to delete saved objects.")
		}

		return jobSavedObjectDeletionResults, http.StatusOK, nil
	}

	// Saved object deletions failed, so we need to fetch the error message from storage
	var metadata shared.ExecutionState
	if err := workflow_utils.ReadFromStorage(
		ctx,
		storageConfig,
		jobMetadataPath,
		&metadata,
	); err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve operator metadata from storage.")
	}

	return emptySavedObjectDeletionResults, http.StatusInternalServerError, errors.New("Unable to delete saved objects.")
}
