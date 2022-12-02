package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/request"
	"github.com/aqueducthq/aqueduct/lib/airflow"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	dag_utils "github.com/aqueducthq/aqueduct/lib/workflow/dag"
	operator_utils "github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// Route: /workflow/register/airflow
// Method: POST
// Params: none
// Request
//	Headers:
//		`api-key`: user's API Key
//	Body:
//		`dag`: a serialized `workflow_dag` object
//		`<operator_id>`: zip file associated with operator for the `operator_id`.
//  	`<operator_id>`: ... (more operator files)
// Response:
//		`file`: a Python file that defines the Airflow DAG

type RegisterAirflowWorkflowHandler struct {
	RegisterWorkflowHandler

	WorkflowDagEdgeReader workflow_dag_edge.Reader

	OperatorResultWriter operator_result.Writer
	ArtifactResultWriter artifact_result.Writer
	NotificationWriter   notification.Writer

	DAGResultRepo repos.DAGResult
}

type registerAirflowWorkflowArgs struct {
	registerWorkflowArgs
}

type registerAirflowWorkflowResponse struct {
	// The newly registered workflow's id.
	Id       uuid.UUID `json:"id"`
	File     string    `json:"file"`
	IsUpdate bool      `json:"is_update"`
}

func (*RegisterAirflowWorkflowHandler) Name() string {
	return "RegisterAirflowWorkflow"
}

func (h *RegisterAirflowWorkflowHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	dagSummary, statusCode, err := request.ParseDagSummaryFromRequest(
		r,
		aqContext.ID,
		h.GithubManager,
		aqContext.StorageConfig,
	)
	if err != nil {
		return nil, statusCode, errors.Wrap(err, "Unable to register workflow.")
	}

	ok, err := dag_utils.ValidateDagOperatorIntegrationOwnership(
		r.Context(),
		dagSummary.Dag.Operators,
		aqContext.OrgID,
		aqContext.ID,
		h.IntegrationReader,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during integration ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own the integrations defined in the Dag.")
	}

	collidingWorkflow, err := h.WorkflowReader.GetWorkflowByName(
		r.Context(),
		dagSummary.Dag.Metadata.UserId,
		dagSummary.Dag.Metadata.Name,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when checking for existing workflows.")
	}

	isUpdate := collidingWorkflow != nil
	if isUpdate {
		// Since the libraries we call use the workflow id to tell whether a workflow already exists.
		dagSummary.Dag.WorkflowID = collidingWorkflow.Id
	}

	if err := dag_utils.Validate(
		dagSummary.Dag,
	); err != nil {
		if _, ok := dag_utils.ValidationErrors[err]; !ok {
			return nil, http.StatusInternalServerError, errors.Wrap(err, "Internal system error occurred while validating the DAG.")
		} else {
			return nil, http.StatusBadRequest, err
		}
	}

	return &registerAirflowWorkflowArgs{
		registerWorkflowArgs: registerWorkflowArgs{
			AqContext:  aqContext,
			dagSummary: dagSummary,
			isUpdate:   isUpdate,
		},
	}, http.StatusOK, nil
}

func (h *RegisterAirflowWorkflowHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*registerAirflowWorkflowArgs)
	dbWorkflowDag := args.dagSummary.Dag
	fileContentsByOperatorID := args.dagSummary.FileContentsByOperatorUUID

	emptyResp := registerAirflowWorkflowResponse{}

	if args.isUpdate {
		// Sync existing Airflow DAGRuns before DAG is updated
		dag, err := utils.ReadLatestDAGFromDatabase(
			ctx,
			dbWorkflowDag.WorkflowID,
			h.WorkflowReader,
			h.DAGRepo,
			h.OperatorReader,
			h.ArtifactReader,
			h.WorkflowDagEdgeReader,
			h.Database,
		)
		if err != nil {
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to update workflow.")
		}

		// NOTE (saurav): This is not perfect because if there are any in progress Airflow DAGRuns, those will
		// not get synced, and may fail to sync later on if the DAG structure has changed.
		if err := airflow.SyncDAGs(
			ctx,
			[]uuid.UUID{dag.ID},
			h.WorkflowReader,
			h.DAGRepo,
			h.OperatorReader,
			h.ArtifactReader,
			h.WorkflowDagEdgeReader,
			h.DAGResultRepo,
			h.OperatorResultWriter,
			h.ArtifactResultWriter,
			h.Vault,
			h.Database,
		); err != nil {
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to update workflow.")
		}
	}

	if _, err := operator_utils.UploadOperatorFiles(ctx, dbWorkflowDag, fileContentsByOperatorID); err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to create workflow.")
	}

	txn, err := h.Database.BeginTx(ctx)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to create workflow.")
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	workflowID, err := utils.WriteDAGToDatabase(
		ctx,
		dbWorkflowDag,
		h.WorkflowReader,
		h.WorkflowWriter,
		h.DAGRepo,
		h.OperatorReader,
		h.OperatorWriter,
		h.WorkflowDagEdgeWriter,
		h.ArtifactReader,
		h.ArtifactWriter,
		txn,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to create workflow.")
	}

	if args.isUpdate {
		// Update workflow metadata and schedule if necessary
		changes := map[string]interface{}{}
		if dbWorkflowDag.Metadata.Name != "" {
			changes[workflow.NameColumn] = dbWorkflowDag.Metadata.Name
		}

		if dbWorkflowDag.Metadata.Description != "" {
			changes[workflow.DescriptionColumn] = dbWorkflowDag.Metadata.Description
		}

		if dbWorkflowDag.Metadata.Schedule.Trigger != "" {
			changes[workflow.ScheduleColumn] = &dbWorkflowDag.Metadata.Schedule
		}

		_, err := h.WorkflowWriter.UpdateWorkflow(ctx, workflowID, changes, txn)
		if err != nil {
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to update workflow.")
		}
	}

	// This is a hack to read the actual operator and artifact IDs generated by the database, since
	// WriteWorkflowDagToDatabase does not update these values.
	dag, err := utils.ReadLatestDAGFromDatabase(
		ctx,
		workflowID,
		h.WorkflowReader,
		h.DAGRepo,
		h.OperatorReader,
		h.ArtifactReader,
		h.WorkflowDagEdgeReader,
		txn,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to create workflow.")
	}

	airflowFile, err := airflow.ScheduleWorkflow(
		ctx,
		dag,
		h.DAGRepo,
		h.JobManager,
		h.Vault,
		txn,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to create workflow.")
	}

	if err := txn.Commit(ctx); err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to create workflow.")
	}

	if !args.isUpdate {
		// Add watcher since this is a new workflow
		watchWorkflowArgs := &watchWorkflowArgs{
			AqContext:  args.AqContext,
			workflowId: workflowID,
		}

		_, _, err = (&WatchWorkflowHandler{
			Database:       h.Database,
			WorkflowReader: h.WorkflowReader,
			WatcherRepo:    h.WatcherRepo,
		}).Perform(ctx, watchWorkflowArgs)
		if err != nil {
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to add user who created the workflow to watch.")
		}
	}

	return &registerAirflowWorkflowResponse{
		Id:       workflowID,
		File:     string(airflowFile),
		IsUpdate: args.isUpdate,
	}, http.StatusOK, nil
}
