package handler

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"encoding/json"
	"time"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_result"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_watcher"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	shared_utils "github.com/aqueducthq/aqueduct/lib/lib_utils"
	workflow_utils "github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

const (
	pollDeleteWrittenObjectsInterval = 500 * time.Millisecond
	pollDeleteWrittenObjectsTimeout  = 2 * time.Minute
)

type TableOutput struct {
	Name string `json:"name"`
	Result bool `json:"succeeded"`
}


// The `DeleteWorkflowHandler` does a best effort at deleting a workflow and its dependencies, such as
// k8s resources, Postgres state, and output tables in the user's data warehouse.
type deleteWorkflowArgs struct {
	*aq_context.AqContext
	WorkflowId uuid.UUID
	ExternalDelete   map[string][]string
	Force   bool
}

type deleteWorkflowInput struct {
	ExternalDelete map[string][]string `json:"external_delete"`
	Force   bool `json:"force"`
}

type deleteWorkflowResponse struct{
	WritesResults map[uuid.UUID][]TableOutput `json:"writes_results"`
}

type DeleteWorkflowHandler struct {
	PostHandler

	Database                database.Database
	StorageConfig     *shared.StorageConfig
	JobManager        job.JobManager
	Vault             vault.Vault
	WorkflowReader          workflow.Reader
	WorkflowDagReader       workflow_dag.Reader
	WorkflowDagEdgeReader   workflow_dag_edge.Reader
	WorkflowDagResultReader workflow_dag_result.Reader
	OperatorReader          operator.Reader
	OperatorResultReader    operator_result.Reader
	ArtifactResultReader    artifact_result.Reader
	IntegrationReader          integration.Reader

	WorkflowWriter          workflow.Writer
	WorkflowDagWriter       workflow_dag.Writer
	WorkflowDagEdgeWriter   workflow_dag_edge.Writer
	WorkflowDagResultWriter workflow_dag_result.Writer
	WorkflowWatcherWriter   workflow_watcher.Writer
	OperatorWriter          operator.Writer
	OperatorResultWriter    operator_result.Writer
	ArtifactWriter          artifact.Writer
	ArtifactResultWriter    artifact_result.Writer
}

func (*DeleteWorkflowHandler) Name() string {
	return "DeleteWorkflow"
}

func (h *DeleteWorkflowHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statuscode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statuscode, err
	}

	workflowIdStr := chi.URLParam(r, routes.WorkflowIdUrlParam)
	workflowId, err := uuid.Parse(workflowIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed workflow ID.")
	}

	ok, err := h.WorkflowReader.ValidateWorkflowOwnership(
		r.Context(),
		workflowId,
		aqContext.OrganizationId,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during workflow ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this workflow.")
	}

	var input deleteWorkflowInput
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return nil, http.StatusBadRequest, errors.New("Unable to parse JSON input.")
	}

	return &deleteWorkflowArgs{
		AqContext:  aqContext,
		WorkflowId: workflowId,
		ExternalDelete:   input.ExternalDelete,
		Force:   input.Force,
	}, http.StatusOK, nil
}

func (h *DeleteWorkflowHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*deleteWorkflowArgs)

	resp := deleteWorkflowResponse{}
	resp.WritesResults = map[uuid.UUID][]TableOutput{}

	// Check tables in list are valid
	objCount := 0
	for integrationId, writeList := range args.ExternalDelete {
		for _, name := range writeList {
			touched, err := h.OperatorReader.TableTouchedByWorkflow(ctx, args.WorkflowId, integrationId, name, h.Database)
			if err != nil {
				return resp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while validating tables.")
			}
			if touched == false {
				return resp, http.StatusBadRequest, errors.Wrap(err, "Table list not valid. Make sure all tables are touched by the workflow.")
			}
			objCount += 1
		}
	}

	// Delete associated tables.
	if objCount > 0 {
		writesResults, httpResponse, err := DeleteWrittenObject(ctx, args, h.Vault, h.StorageConfig, h.JobManager, h.Database, h.IntegrationReader)
		if httpResponse != http.StatusOK {
			return resp, httpResponse, err
		}
		resp.WritesResults = writesResults
	}

	txn, err := h.Database.BeginTx(ctx)
	if err != nil {
		return resp, http.StatusInternalServerError, errors.Wrap(err, "Unable to delete workflow.")
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	// We first retrieve all relevant records from the database.
	workflowObject, err := h.WorkflowReader.GetWorkflow(ctx, args.WorkflowId, txn)
	if err != nil {
		return resp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving workflow.")
	}

	workflowDagsToDelete, err := h.WorkflowDagReader.GetWorkflowDagsByWorkflowId(ctx, workflowObject.Id, txn)
	if err != nil || len(workflowDagsToDelete) == 0 {
		return resp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving workflow dags.")
	}

	workflowDagIds := make([]uuid.UUID, 0, len(workflowDagsToDelete))
	for _, workflowDag := range workflowDagsToDelete {
		workflowDagIds = append(workflowDagIds, workflowDag.Id)
	}

	workflowDagResultsToDelete, err := h.WorkflowDagResultReader.GetWorkflowDagResultsByWorkflowId(ctx, workflowObject.Id, txn)
	if err != nil {
		return resp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving workflow dag results.")
	}

	workflowDagResultIds := make([]uuid.UUID, 0, len(workflowDagResultsToDelete))
	for _, workflowDagResult := range workflowDagResultsToDelete {
		workflowDagResultIds = append(workflowDagResultIds, workflowDagResult.Id)
	}

	workflowDagEdgesToDelete, err := h.WorkflowDagEdgeReader.GetEdgesByWorkflowDagIds(ctx, workflowDagIds, txn)
	if err != nil {
		return resp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving workflow dag edges.")
	}

	operatorIds := make([]uuid.UUID, 0, len(workflowDagEdgesToDelete))
	artifactIds := make([]uuid.UUID, 0, len(workflowDagEdgesToDelete))

	operatorIdMap := make(map[uuid.UUID]bool)
	artifactIdMap := make(map[uuid.UUID]bool)

	for _, workflowDagEdge := range workflowDagEdgesToDelete {
		var operatorId uuid.UUID
		var artifactId uuid.UUID

		if workflowDagEdge.Type == workflow_dag_edge.OperatorToArtifactType {
			operatorId = workflowDagEdge.FromId
			artifactId = workflowDagEdge.ToId
		} else {
			operatorId = workflowDagEdge.ToId
			artifactId = workflowDagEdge.FromId
		}

		if _, ok := operatorIdMap[operatorId]; !ok {
			operatorIdMap[operatorId] = true
			operatorIds = append(operatorIds, operatorId)
		}

		if _, ok := artifactIdMap[artifactId]; !ok {
			artifactIdMap[artifactId] = true
			artifactIds = append(artifactIds, artifactId)
		}
	}

	operatorsToDelete, err := h.OperatorReader.GetOperators(ctx, operatorIds, txn)
	if err != nil {
		return resp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving operators.")
	}

	operatorResultsToDelete, err := h.OperatorResultReader.GetOperatorResultsByWorkflowDagResultIds(
		ctx,
		workflowDagResultIds,
		txn,
	)
	if err != nil {
		return resp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving operator results.")
	}

	operatorResultIds := make([]uuid.UUID, 0, len(operatorResultsToDelete))
	for _, operatorResult := range operatorResultsToDelete {
		operatorResultIds = append(operatorResultIds, operatorResult.Id)
	}

	artifactResultsToDelete, err := h.ArtifactResultReader.GetArtifactResultsByWorkflowDagResultIds(
		ctx,
		workflowDagResultIds,
		txn,
	)
	if err != nil {
		return resp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving artifact results.")
	}

	artifactResultIds := make([]uuid.UUID, 0, len(artifactResultsToDelete))
	for _, artifactResult := range artifactResultsToDelete {
		artifactResultIds = append(artifactResultIds, artifactResult.Id)
	}

	// Start deleting database records.
	err = h.WorkflowWatcherWriter.DeleteWorkflowWatcherByWorkflowId(ctx, workflowObject.Id, txn)
	if err != nil {
		return resp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting workflow watchers.")
	}

	err = h.OperatorResultWriter.DeleteOperatorResults(ctx, operatorResultIds, txn)
	if err != nil {
		return resp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting operator results.")
	}

	err = h.ArtifactResultWriter.DeleteArtifactResults(ctx, artifactResultIds, txn)
	if err != nil {
		return resp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting artifact results.")
	}

	err = h.WorkflowDagResultWriter.DeleteWorkflowDagResults(ctx, workflowDagResultIds, txn)
	if err != nil {
		return resp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting workflow dag results.")
	}

	err = h.WorkflowDagEdgeWriter.DeleteEdgesByWorkflowDagIds(ctx, workflowDagIds, txn)
	if err != nil {
		return resp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting workflow dag edges.")
	}

	err = h.OperatorWriter.DeleteOperators(ctx, operatorIds, txn)
	if err != nil {
		return resp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting operators.")
	}

	err = h.ArtifactWriter.DeleteArtifacts(ctx, artifactIds, txn)
	if err != nil {
		return resp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting artifacts.")
	}

	err = h.WorkflowDagWriter.DeleteWorkflowDags(ctx, workflowDagIds, txn)
	if err != nil {
		return resp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting workflow dags.")
	}

	err = h.WorkflowWriter.DeleteWorkflow(ctx, workflowObject.Id, txn)
	if err != nil {
		return resp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting workflow.")
	}

	if err := txn.Commit(ctx); err != nil {
		return resp, http.StatusInternalServerError, errors.Wrap(err, "Failed to delete workflow.")
	}

	// Delete storage files (artifact content and function files)
	storagePaths := make([]string, 0, len(operatorIds)+len(artifactResultIds))
	for _, op := range operatorsToDelete {
		if op.Spec.IsFunction() {
			storagePaths = append(storagePaths, op.Spec.Function().StoragePath)
		}
	}

	for _, art := range artifactResultsToDelete {
		storagePaths = append(storagePaths, art.ContentPath)
	}

	// Note: for now we assume all workflow dags have the same storage config.
	// This assumption will stay true until we allow users to configure custom storage config to store stuff.
	storageConfig := workflowDagsToDelete[0].StorageConfig
	for _, workflowDag := range workflowDagsToDelete {
		if !reflect.DeepEqual(workflowDag.StorageConfig, storageConfig) {
			return resp, http.StatusInternalServerError, errors.New("Workflow Dags have mismatching storage config.")
		}
	}

	workflow_utils.CleanupStorageFiles(ctx, &storageConfig, storagePaths)

	// Delete the cron job if it had one.
	if workflowObject.Schedule.CronSchedule != "" {
		cronjobName := shared_utils.AppendPrefix(workflowObject.Id.String())
		err = h.JobManager.DeleteCronJob(ctx, cronjobName)
		if err != nil {
			return resp, http.StatusInternalServerError, errors.Wrap(err, "Failed to delete workflow's cronjob.")
		}
	}

	return resp, http.StatusOK, nil
}

func DeleteWrittenObject(ctx context.Context, args *deleteWorkflowArgs, vaultObject vault.Vault, storageConfig *shared.StorageConfig, jobManager job.JobManager, db database.Database, intergrationReader integration.Reader) (map[uuid.UUID][]TableOutput, int, error) {
	emptyWritesResults := map[uuid.UUID][]TableOutput{}

	// Schedule delete written objects job
	jobMetadataPath := fmt.Sprintf("delete-written-objects-%s", args.RequestId)

	jobName := fmt.Sprintf("delete-written-objects-%s", uuid.New().String())
	contentPath := fmt.Sprintf("delete-written-objects-content-%s", args.RequestId)

	integrationConfigs := map[string]auth.Config{}
	integrationNames := map[string]integration.Service{}
	for integrationId, _ := range args.ExternalDelete {
		integrationUUID, err := uuid.Parse(integrationId)
		if err != nil {
			return emptyWritesResults, http.StatusInternalServerError, errors.Wrap(err, "Unable to get integration configs.")
		}
		config, err := auth.ReadConfigFromSecret(ctx, integrationUUID, vaultObject)
		if err != nil {
			return emptyWritesResults, http.StatusInternalServerError, errors.Wrap(err, "Unable to get integration configs.")
		}
		integrationConfigs[integrationId] = config
		integrationObjects, err := intergrationReader.GetIntegrations(ctx, []uuid.UUID{integrationUUID}, db)
		if err != nil {
			return emptyWritesResults, http.StatusInternalServerError, errors.Wrap(err, "Unable to get integration configs.")
		}
		if len(integrationObjects) != 1 {
			return emptyWritesResults, http.StatusInternalServerError, errors.Wrap(err, "Unable to get integration configs.")
		}
		integrationNames[integrationId] = integrationObjects[0].Service
	}

	jobSpec := job.NewDeleteWrittenObjectsSpec(
		jobName,
		storageConfig,
		jobMetadataPath,
		integrationNames,
		integrationConfigs,
		args.ExternalDelete,
		contentPath,
	)
	if err := jobManager.Launch(ctx, jobName, jobSpec); err != nil {
		return emptyWritesResults, http.StatusInternalServerError, errors.Wrap(err, "Unable to launch delete written objects job.")
	}

	jobStatus, err := job.PollJob(ctx, jobName, jobManager, pollDeleteWrittenObjectsInterval, pollDeleteWrittenObjectsTimeout)
	if err != nil {
		return emptyWritesResults, http.StatusInternalServerError, errors.Wrap(err, "Unable to delete written objects.")
	}

	if jobStatus == shared.SucceededExecutionStatus {
		// Table deletions were successful
		jobWritesResults := map[uuid.UUID][]TableOutput{}

		if err := workflow_utils.ReadFromStorage(
			ctx,
			storageConfig,
			contentPath,
			&jobWritesResults,
		); err != nil {
			return emptyWritesResults, http.StatusInternalServerError, errors.Wrap(err, "Unable to delete written objects.")
		}

		return jobWritesResults, http.StatusOK, nil
	}

	// Written object deletions failed, so we need to fetch the error message from storage
	var metadata shared.ExecutionState
	if err := workflow_utils.ReadFromStorage(
		ctx,
		storageConfig,
		jobMetadataPath,
		&metadata,
	); err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve operator metadata from storage.")
	}


	return emptyWritesResults, http.StatusInternalServerError, errors.New("Unable to delete written objects.")
}
