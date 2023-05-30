package v2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/config"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	exec_env "github.com/aqueducthq/aqueduct/lib/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/job"
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

// Route:
//
//	v2/workflow/{workflowId}/delete
//	workflow/{workflowId}/delete
//
// Method: POST
// Params: workflowId
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//	Body:
//		json-serialized `workflowDeleteInput` object.
//
// Response: json-serialized `workflowDeleteResponse` object.
//
// The `WorkflowDeleteHandler` does a best effort at deleting a workflow and its dependencies, such as
// k8s resources, Postgres state, and output objects in the user's data warehouse.
type workflowDeleteArgs struct {
	*aq_context.AqContext
	WorkflowID     uuid.UUID
	ExternalDelete map[string][]string
	Force          bool
}

type workflowDeleteInput struct {
	// This is a map from resource_id to the serialized load spec we want to delete.
	ExternalDeleteLoadParams map[string][]string `json:"external_delete"`
	// `Force` serve as a safe-guard for client to confirm the deletion.
	// If `Force` is true, all objects specified in `ExternalDelete` field
	// will be removed. Otherwise, we will not delete the objects.
	Force bool `json:"force"`
}

type workflowDeleteResponse struct {
	// This is a map from resource_id to a list of `SavedObjectResult`
	// implying if each object is successfully deleted.
	SavedObjectDeletionResults map[string][]SavedObjectResult `json:"saved_object_deletion_results"`
}

type WorkflowDeleteHandler struct {
	handler.PostHandler

	Database   database.Database
	Engine     engine.Engine
	JobManager job.JobManager

	ExecutionEnvironmentRepo repos.ExecutionEnvironment
	ResourceRepo             repos.Resource
	OperatorRepo             repos.Operator
	WorkflowRepo             repos.Workflow
	DagRepo                  repos.DAG
	ArtifactResultRepo       repos.ArtifactResult
}

func (*WorkflowDeleteHandler) Name() string {
	return "WorkflowDelete"
}

func (h *WorkflowDeleteHandler) Prepare(r *http.Request) (interface{}, int, error) {
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

	var input workflowDeleteInput
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to parse JSON input.")
	}

	// Convert the supplied load params into object identifiers (eg. object names for relational databases)
	externalDelete := make(map[string][]string, len(input.ExternalDeleteLoadParams))
	for resourceName, loadSpecStrList := range input.ExternalDeleteLoadParams {
		for _, loadSpecStr := range loadSpecStrList {
			var loadSpec connector.Load
			err = json.Unmarshal([]byte(loadSpecStr), &loadSpec)
			if err != nil {
				return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to parse request.")
			}

			if relationalLoadParams, ok := connector.CastToRelationalDBLoadParams(loadSpec.Parameters); ok {
				externalDelete[resourceName] = append(externalDelete[resourceName], relationalLoadParams.Table)
			} else if s3LoadParams, ok := loadSpec.Parameters.(*connector.S3LoadParams); ok {
				externalDelete[resourceName] = append(externalDelete[resourceName], s3LoadParams.Filepath)
			} else {
				return nil, http.StatusBadRequest, errors.Newf("Unsupported resource type for deleting objects: %s", resourceName)
			}
		}
	}

	return &workflowDeleteArgs{
		AqContext:      aqContext,
		WorkflowID:     workflowID,
		ExternalDelete: externalDelete,
		Force:          input.Force,
	}, http.StatusOK, nil
}

func (h *WorkflowDeleteHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*workflowDeleteArgs)

	resp := workflowDeleteResponse{}
	resp.SavedObjectDeletionResults = map[string][]SavedObjectResult{}

	nameToID := make(map[string]uuid.UUID, len(args.ExternalDelete))
	for resourceName := range args.ExternalDelete {
		resourceObject, err := h.ResourceRepo.GetByNameAndUser(
			ctx,
			resourceName,
			args.AqContext.ID,
			args.AqContext.OrgID,
			h.Database,
		)
		if err != nil {
			return resp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while getting resource.")
		}
		nameToID[resourceName] = resourceObject.ID
	}

	// Check objects in list are valid
	objCount := 0

	// These fetched save operators have any parameterized values filled in.
	saveOpsByResourceName := make(map[string][]views.LoadOperator, 1)
	saveOpsList, err := GetDistinctLoadOpsByWorkflow(
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
		saveOpsByResourceName[saveOp.ResourceName] = append(saveOpsByResourceName[saveOp.ResourceName], saveOpsList...)
	}

	for resourceName, savedObjectList := range args.ExternalDelete {
		for _, name := range savedObjectList {
			var touchedOperatorLoadParams []connector.LoadParams

			// Check for existence in the parameter-expanded list.
			for _, saveOp := range saveOpsByResourceName[resourceName] {
				relationalLoad, ok := connector.CastToRelationalDBLoadParams(saveOp.Spec.Parameters)
				if ok && relationalLoad.Table == name {
					touchedOperatorLoadParams = append(touchedOperatorLoadParams, saveOp.Spec.Parameters)
				}

				nonRelationalLoad, ok := connector.CastToNonRelationalLoadParams(saveOp.Spec.Parameters)
				if ok && nonRelationalLoad.Filepath == name {
					touchedOperatorLoadParams = append(touchedOperatorLoadParams, saveOp.Spec.Parameters)
				}
			}

			if len(touchedOperatorLoadParams) == 0 {
				return resp, http.StatusBadRequest, errors.New("Object list not valid. Make sure all objects are touched by the workflow.")
			}

			if !args.Force {
				// Check none have UpdateMode=append.
				for _, touchedLoadParams := range touchedOperatorLoadParams {
					relationalLoad, ok := connector.CastToRelationalDBLoadParams(touchedLoadParams)
					// Check not updating anything in the resource.
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
			h.ResourceRepo,
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
	args *workflowDeleteArgs,
	resourceNameToID map[string]uuid.UUID,
	vaultObject vault.Vault,
	storageConfig *shared.StorageConfig,
	jobManager job.JobManager,
	DB database.Database,
	resourceRepo repos.Resource,
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

	resourceConfigs := make(map[string]auth.Config, len(resourceNameToID))
	resourceNames := make(map[string]shared.Service, len(resourceNameToID))
	for resourceName := range args.ExternalDelete {
		resourceId := resourceNameToID[resourceName]
		config, err := auth.ReadConfigFromSecret(ctx, resourceId, vaultObject)
		if err != nil {
			return emptySavedObjectDeletionResults, http.StatusInternalServerError, errors.Wrap(err, "Unable to get resource configs.")
		}
		resourceConfigs[resourceName] = config
		resourceObjects, err := resourceRepo.GetBatch(ctx, []uuid.UUID{resourceId}, DB)
		if err != nil {
			return emptySavedObjectDeletionResults, http.StatusInternalServerError, errors.Wrap(err, "Unable to get resource configs.")
		}
		if len(resourceObjects) != 1 {
			return emptySavedObjectDeletionResults, http.StatusInternalServerError, errors.New("Unable to get resource configs.")
		}
		resourceNames[resourceName] = resourceObjects[0].Service
	}

	jobSpec := job.NewDeleteSavedObjectsSpec(
		jobName,
		storageConfig,
		jobMetadataPath,
		resourceNames,
		resourceConfigs,
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
