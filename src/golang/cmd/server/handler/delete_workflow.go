package handler

import (
	"context"
	"net/http"
	// "reflect"
	"encoding/json"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	// "github.com/aqueducthq/aqueduct/lib/collections/operator/connector"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_result"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_watcher"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	// shared_utils "github.com/aqueducthq/aqueduct/lib/lib_utils"
	// workflow_utils "github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type TableOutput struct {
	name string `json:"name"`
	result bool `json:"succeeded"`
}


// The `DeleteWorkflowHandler` does a best effort at deleting a workflow and its dependencies, such as
// k8s resources, Postgres state, and output tables in the user's data warehouse.
type deleteWorkflowArgs struct {
	*aq_context.AqContext
	workflowId uuid.UUID
	externalDelete   map[uuid.UUID][]str `json:"external_delete"`
	force   bool `json:"force"`
}

type deleteWorkflowInput struct {
	externalDelete map[uuid.UUID][]str `json:"external_delete"`
	force   bool `json:"force"`
}

type deleteWorkflowResponse struct{
	writesResults map[uuid.UUID][]TableOutput `json:"writes_results"`
}

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

	fmt.print(input)

	return &deleteWorkflowArgs{
		AqContext:  aqContext,
		workflowId: workflowId,
		externalDelete:   input.externalDelete,
		force:   input.force,
	}, http.StatusOK, nil
}

func (h *DeleteWorkflowHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	// args := interfaceArgs.(*deleteWorkflowArgs)

	emptyResp := deleteWorkflowResponse{}

	// Check tables in list are valid
	// for _, spec := range args.loadSpec {
	// 	relationalParam := connector.CastToRelationalDBLoadParams(spec.Parameters)

	// 	integrations, err := h.IntegrationReader.GetIntegrationsByServiceAndUser(ctx, spec.ConnectorName, args.AqContext.UserId, h.Database)
	// 	if err {
	// 		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving integration id.")
	// 	}

	// 	integrationId := nil
	// 	for _, integration := range integrations {
	// 		eq := reflect.DeepEqual(integration.Config, spec.ConnectorConfig)
	// 		if eq {
	// 			if integrationId == nil {
	// 				integrationId = integration.Id
	// 			} else {
	// 				return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpectedly retrieved multiple integration ids.")
	// 			}
	// 		}
	// 	}
	// 	if integrationId == nil {
	// 		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Could not find integration id.")
	// 	}

	// 	touched, err := h.OperatorReader.TableTouchedByWorkflow(ctx, args.workflowId, integrationId, relationalParam.Table, h.Database)
	// 	if err != nil {
	// 		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while validating tables.")
	// 	}
	// 	if touched == false {
	// 		return emptyResp, http.StatusBadRequest, errors.Wrap(err, "Table list not valid. Make sure all tables are touched by the workflow")
	// 	}
	// }

	// Delete associated tables.
	// tableResults, httpResponse, err := DeleteTable(ctx, args, tableSpecs LoadSpec, integrationObject *integration.Integration, vaultObject vault.Vault, jobManager job.JobManager) ([]TableOutput, int, error)

	// txn, err := h.Database.BeginTx(ctx)
	// if err != nil {
	// 	return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to delete workflow.")
	// }
	// defer database.TxnRollbackIgnoreErr(ctx, txn)

	// // We first retrieve all relevant records from the database.
	// workflowObject, err := h.WorkflowReader.GetWorkflow(ctx, args.workflowId, txn)
	// if err != nil {
	// 	return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving workflow.")
	// }

	// workflowDagsToDelete, err := h.WorkflowDagReader.GetWorkflowDagsByWorkflowId(ctx, workflowObject.Id, txn)
	// if err != nil || len(workflowDagsToDelete) == 0 {
	// 	return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving workflow dags.")
	// }

	// workflowDagIds := make([]uuid.UUID, 0, len(workflowDagsToDelete))
	// for _, workflowDag := range workflowDagsToDelete {
	// 	workflowDagIds = append(workflowDagIds, workflowDag.Id)
	// }

	// workflowDagResultsToDelete, err := h.WorkflowDagResultReader.GetWorkflowDagResultsByWorkflowId(ctx, workflowObject.Id, txn)
	// if err != nil {
	// 	return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving workflow dag results.")
	// }

	// workflowDagResultIds := make([]uuid.UUID, 0, len(workflowDagResultsToDelete))
	// for _, workflowDagResult := range workflowDagResultsToDelete {
	// 	workflowDagResultIds = append(workflowDagResultIds, workflowDagResult.Id)
	// }

	// workflowDagEdgesToDelete, err := h.WorkflowDagEdgeReader.GetEdgesByWorkflowDagIds(ctx, workflowDagIds, txn)
	// if err != nil {
	// 	return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving workflow dag edges.")
	// }

	// operatorIds := make([]uuid.UUID, 0, len(workflowDagEdgesToDelete))
	// artifactIds := make([]uuid.UUID, 0, len(workflowDagEdgesToDelete))

	// operatorIdMap := make(map[uuid.UUID]bool)
	// artifactIdMap := make(map[uuid.UUID]bool)

	// for _, workflowDagEdge := range workflowDagEdgesToDelete {
	// 	var operatorId uuid.UUID
	// 	var artifactId uuid.UUID

	// 	if workflowDagEdge.Type == workflow_dag_edge.OperatorToArtifactType {
	// 		operatorId = workflowDagEdge.FromId
	// 		artifactId = workflowDagEdge.ToId
	// 	} else {
	// 		operatorId = workflowDagEdge.ToId
	// 		artifactId = workflowDagEdge.FromId
	// 	}

	// 	if _, ok := operatorIdMap[operatorId]; !ok {
	// 		operatorIdMap[operatorId] = true
	// 		operatorIds = append(operatorIds, operatorId)
	// 	}

	// 	if _, ok := artifactIdMap[artifactId]; !ok {
	// 		artifactIdMap[artifactId] = true
	// 		artifactIds = append(artifactIds, artifactId)
	// 	}
	// }

	// operatorsToDelete, err := h.OperatorReader.GetOperators(ctx, operatorIds, txn)
	// if err != nil {
	// 	return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving operators.")
	// }

	// operatorResultsToDelete, err := h.OperatorResultReader.GetOperatorResultsByWorkflowDagResultIds(
	// 	ctx,
	// 	workflowDagResultIds,
	// 	txn,
	// )
	// if err != nil {
	// 	return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving operator results.")
	// }

	// operatorResultIds := make([]uuid.UUID, 0, len(operatorResultsToDelete))
	// for _, operatorResult := range operatorResultsToDelete {
	// 	operatorResultIds = append(operatorResultIds, operatorResult.Id)
	// }

	// artifactResultsToDelete, err := h.ArtifactResultReader.GetArtifactResultsByWorkflowDagResultIds(
	// 	ctx,
	// 	workflowDagResultIds,
	// 	txn,
	// )
	// if err != nil {
	// 	return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving artifact results.")
	// }

	// artifactResultIds := make([]uuid.UUID, 0, len(artifactResultsToDelete))
	// for _, artifactResult := range artifactResultsToDelete {
	// 	artifactResultIds = append(artifactResultIds, artifactResult.Id)
	// }

	// // Start deleting database records.
	// err = h.WorkflowWatcherWriter.DeleteWorkflowWatcherByWorkflowId(ctx, workflowObject.Id, txn)
	// if err != nil {
	// 	return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting workflow watchers.")
	// }

	// err = h.OperatorResultWriter.DeleteOperatorResults(ctx, operatorResultIds, txn)
	// if err != nil {
	// 	return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting operator results.")
	// }

	// err = h.ArtifactResultWriter.DeleteArtifactResults(ctx, artifactResultIds, txn)
	// if err != nil {
	// 	return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting artifact results.")
	// }

	// err = h.WorkflowDagResultWriter.DeleteWorkflowDagResults(ctx, workflowDagResultIds, txn)
	// if err != nil {
	// 	return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting workflow dag results.")
	// }

	// err = h.WorkflowDagEdgeWriter.DeleteEdgesByWorkflowDagIds(ctx, workflowDagIds, txn)
	// if err != nil {
	// 	return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting workflow dag edges.")
	// }

	// err = h.OperatorWriter.DeleteOperators(ctx, operatorIds, txn)
	// if err != nil {
	// 	return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting operators.")
	// }

	// err = h.ArtifactWriter.DeleteArtifacts(ctx, artifactIds, txn)
	// if err != nil {
	// 	return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting artifacts.")
	// }

	// err = h.WorkflowDagWriter.DeleteWorkflowDags(ctx, workflowDagIds, txn)
	// if err != nil {
	// 	return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting workflow dags.")
	// }

	// err = h.WorkflowWriter.DeleteWorkflow(ctx, workflowObject.Id, txn)
	// if err != nil {
	// 	return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting workflow.")
	// }

	// if err := txn.Commit(ctx); err != nil {
	// 	return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to delete workflow.")
	// }

	// // Delete storage files (artifact content and function files)
	// storagePaths := make([]string, 0, len(operatorIds)+len(artifactResultIds))
	// for _, op := range operatorsToDelete {
	// 	if op.Spec.IsFunction() {
	// 		storagePaths = append(storagePaths, op.Spec.Function().StoragePath)
	// 	}
	// }

	// for _, art := range artifactResultsToDelete {
	// 	storagePaths = append(storagePaths, art.ContentPath)
	// }

	// // Note: for now we assume all workflow dags have the same storage config.
	// // This assumption will stay true until we allow users to configure custom storage config to store stuff.
	// storageConfig := workflowDagsToDelete[0].StorageConfig
	// for _, workflowDag := range workflowDagsToDelete {
	// 	if !reflect.DeepEqual(workflowDag.StorageConfig, storageConfig) {
	// 		return emptyResp, http.StatusInternalServerError, errors.New("Workflow Dags have mismatching storage config.")
	// 	}
	// }

	// workflow_utils.CleanupStorageFiles(ctx, &storageConfig, storagePaths)

	// // Delete the cron job if it had one.
	// if workflowObject.Schedule.CronSchedule != "" {
	// 	cronjobName := shared_utils.AppendPrefix(workflowObject.Id.String())
	// 	err = h.JobManager.DeleteCronJob(ctx, cronjobName)
	// 	if err != nil {
	// 		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to delete workflow's cronjob.")
	// 	}
	// }

	return emptyResp, http.StatusOK, nil
}

// func DeleteTable(ctx context.Context, args *DeleteTableArgs, tableSpecs LoadSpec, integrationObject *integration.Integration, vaultObject vault.Vault, jobManager job.JobManager) (int, error) {
// 	emptyTables := make([]TableOutput, 0)

// 	// Schedule delete table job
// 	jobMetadataPath := fmt.Sprintf("delete-tables-%s", args.RequestId)

// 	jobName := fmt.Sprintf("delete-tables-operator-%s", uuid.New().String())

// 	config, err := auth.ReadConfigFromSecret(ctx, integrationObject.Id, vaultObject)
// 	if err != nil {
// 		return http.StatusInternalServerError, errors.Wrap(err, "Unable to launch delete tables job.")
// 	}

// 	jobSpec := job.NewDeleteTablesSpec(
// 		jobName,
// 		storageConfig,
// 		jobMetadataPath,
// 		integrationObject.Service,
// 		config,
// 		loadParameters,
// 		contentPath,
// 	)
// 	if err := jobManager.Launch(ctx, jobName, jobSpec); err != nil {
// 		return http.StatusInternalServerError, errors.Wrap(err, "Unable to launch delete tables job.")
// 	}

// 	jobStatus, err := job.PollJob(ctx, jobName, jobManager, pollCreateInterval, pollCreateTimeout)
// 	if err != nil {
// 		return http.StatusInternalServerError, errors.Wrap(err, "Unable to delete tables.")
// 	}

// 	if jobStatus == shared.SucceededExecutionStatus {
// 		// Table deletions were successful
// 		var tables []TableOutput
// 		if err := utils.ReadFromStorage(
// 			ctx,
// 			storageConfig,
// 			contentPath,
// 			&tables,
// 		); err != nil {
// 			return http.StatusInternalServerError, errors.Wrap(err, "Unable to delete tables.")
// 		}
// 		return tables, http.StatusOK, nil
// 	}

// 	// Table deletions failed, so we need to fetch the error message from storage
// 	var metadata operator_result.Metadata
// 	if err := utils.ReadFromStorage(
// 		ctx,
// 		storageConfig,
// 		jobMetadataPath,
// 		&metadata,
// 	); err != nil {
// 		return http.StatusInternalServerError, errors.Wrap(err, "Unable to delete tables.")
// 	}

// 	return http.StatusBadRequest, errors.Newf("Unable to delete tables: %v", metadata.Error)
// }
