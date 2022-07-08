package handler

import (
	"context"
	"net/http"
	"reflect"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_result"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_watcher"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	shared_utils "github.com/aqueducthq/aqueduct/lib/lib_utils"
	workflow_utils "github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

// The `DeleteWorkflowHandler` does a best effort at deleting a workflow and its dependencies, such as
// k8s resources, Postgres state, and output tables in the user's data warehouse.
type deleteWorkflowArgs struct {
	*aq_context.AqContext
	workflowId uuid.UUID
	tables []string
}

type deleteWorkflowInput struct {
	tables        []string             `json:"tables"`
}


type deleteWorkflowResponse struct{}

type DeleteWorkflowHandler struct {
	PostHandler

	Database                database.Database
	JobManager              job.JobManager
	WorkflowReader          workflow.Reader
	WorkflowDagReader       workflow_dag.Reader
	WorkflowDagEdgeReader   workflow_dag_edge.Reader
	WorkflowDagResultReader workflow_dag_result.Reader
	OperatorReader          operator.Reader
	OperatorResultReader    operator_result.Reader
	ArtifactResultReader    artifact_result.Reader

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

	var input deleteWorkflowInput 
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return nil, http.StatusBadRequest, errors.New("Unable to parse JSON input.")
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

	return &deleteWorkflowArgs{
		AqContext:  aqContext,
		workflowId: workflowId,
		tables:		input.tables,
	}, http.StatusOK, nil
}

func (h *DeleteWorkflowHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {

	// TODO: Check each table is associated with the workflow. Else, return early.

	// TODO: Delete associated tables.

	args := interfaceArgs.(*deleteWorkflowArgs)

	emptyResp := deleteWorkflowResponse{}

	txn, err := h.Database.BeginTx(ctx)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to delete workflow.")
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	// We first retrieve all relevant records from the database.
	workflowObject, err := h.WorkflowReader.GetWorkflow(ctx, args.workflowId, txn)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving workflow.")
	}

	workflowDagsToDelete, err := h.WorkflowDagReader.GetWorkflowDagsByWorkflowId(ctx, workflowObject.Id, txn)
	if err != nil || len(workflowDagsToDelete) == 0 {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving workflow dags.")
	}

	workflowDagIds := make([]uuid.UUID, 0, len(workflowDagsToDelete))
	for _, workflowDag := range workflowDagsToDelete {
		workflowDagIds = append(workflowDagIds, workflowDag.Id)
	}

	workflowDagResultsToDelete, err := h.WorkflowDagResultReader.GetWorkflowDagResultsByWorkflowId(ctx, workflowObject.Id, txn)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving workflow dag results.")
	}

	workflowDagResultIds := make([]uuid.UUID, 0, len(workflowDagResultsToDelete))
	for _, workflowDagResult := range workflowDagResultsToDelete {
		workflowDagResultIds = append(workflowDagResultIds, workflowDagResult.Id)
	}

	workflowDagEdgesToDelete, err := h.WorkflowDagEdgeReader.GetEdgesByWorkflowDagIds(ctx, workflowDagIds, txn)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving workflow dag edges.")
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
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving operators.")
	}

	operatorResultsToDelete, err := h.OperatorResultReader.GetOperatorResultsByWorkflowDagResultIds(
		ctx,
		workflowDagResultIds,
		txn,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving operator results.")
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
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving artifact results.")
	}

	artifactResultIds := make([]uuid.UUID, 0, len(artifactResultsToDelete))
	for _, artifactResult := range artifactResultsToDelete {
		artifactResultIds = append(artifactResultIds, artifactResult.Id)
	}

	// Start deleting database records.
	err = h.WorkflowWatcherWriter.DeleteWorkflowWatcherByWorkflowId(ctx, workflowObject.Id, txn)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting workflow watchers.")
	}

	err = h.OperatorResultWriter.DeleteOperatorResults(ctx, operatorResultIds, txn)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting operator results.")
	}

	err = h.ArtifactResultWriter.DeleteArtifactResults(ctx, artifactResultIds, txn)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting artifact results.")
	}

	err = h.WorkflowDagResultWriter.DeleteWorkflowDagResults(ctx, workflowDagResultIds, txn)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting workflow dag results.")
	}

	err = h.WorkflowDagEdgeWriter.DeleteEdgesByWorkflowDagIds(ctx, workflowDagIds, txn)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting workflow dag edges.")
	}

	err = h.OperatorWriter.DeleteOperators(ctx, operatorIds, txn)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting operators.")
	}

	err = h.ArtifactWriter.DeleteArtifacts(ctx, artifactIds, txn)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting artifacts.")
	}

	err = h.WorkflowDagWriter.DeleteWorkflowDags(ctx, workflowDagIds, txn)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting workflow dags.")
	}

	err = h.WorkflowWriter.DeleteWorkflow(ctx, workflowObject.Id, txn)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting workflow.")
	}

	if err := txn.Commit(ctx); err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to delete workflow.")
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
			return emptyResp, http.StatusInternalServerError, errors.New("Workflow Dags have mismatching storage config.")
		}
	}

	workflow_utils.CleanupStorageFiles(ctx, &storageConfig, storagePaths)

	// Delete the cron job if it had one.
	if workflowObject.Schedule.CronSchedule != "" {
		cronjobName := shared_utils.AppendPrefix(workflowObject.Id.String())
		err = h.JobManager.DeleteCronJob(ctx, cronjobName)
		if err != nil {
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to delete workflow's cronjob.")
		}
	}

	return emptyResp, http.StatusOK, nil
}
