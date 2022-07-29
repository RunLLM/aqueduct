package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/request"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/user"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_result"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_watcher"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	shared_utils "github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/vault"
	dag_utils "github.com/aqueducthq/aqueduct/lib/workflow/dag"
	"github.com/aqueducthq/aqueduct/lib/workflow/engine"
	operator_utils "github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/github"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// Route: /workflow/register
// Method: POST
// Params: none
// Request
//	Headers:
//		`api-key`: user's API Key
//	Body:
//		`dag`: a serialized `workflow_dag` object
//		`<operator_id>`: zip file associated with operator for the `operator_id`.
//  	`<operator_id>`: ... (more operator files)
// Response: none

type RegisterWorkflowHandler struct {
	PostHandler

	Database      database.Database
	JobManager    job.JobManager
	GithubManager github.Manager
	Vault         vault.Vault
	StorageConfig *shared.StorageConfig

	ArtifactReader    artifact.Reader
	IntegrationReader integration.Reader
	OperatorReader    operator.Reader
	UserReader        user.Reader
	WorkflowReader    workflow.Reader

	ArtifactWriter          artifact.Writer
	ArtifactResultWriter    artifact_result.Writer
	NotificationWriter      notification.Writer
	OperatorWriter          operator.Writer
	OperatorResultWriter    operator_result.Writer
	WorkflowWriter          workflow.Writer
	WorkflowDagWriter       workflow_dag.Writer
	WorkflowDagEdgeWriter   workflow_dag_edge.Writer
	WorkflowDagResultWriter workflow_dag_result.Writer
	WorkflowWatcherWriter   workflow_watcher.Writer
}

type registerWorkflowArgs struct {
	*aq_context.AqContext
	dbWorkflowDag            *workflow_dag.DBWorkflowDag
	operatorIdToFileContents map[uuid.UUID][]byte

	// Whether this is a registering a new workflow or updating an existing one.
	isUpdate bool
}

type registerWorkflowResponse struct {
	// The newly registered workflow's id.
	Id uuid.UUID `json:"id"`
}

func (*RegisterWorkflowHandler) Name() string {
	return "RegisterWorkflow"
}

func (h *RegisterWorkflowHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	dagSummary, statusCode, err := request.ParseDagSummaryFromRequest(
		r,
		aqContext.Id,
		h.GithubManager,
		h.StorageConfig,
	)
	if err != nil {
		return nil, statusCode, errors.Wrap(err, "Unable to register workflow.")
	}

	ok, err := dag_utils.ValidateDagOperatorIntegrationOwnership(
		r.Context(),
		dagSummary.Dag.Operators,
		aqContext.OrganizationId,
		h.IntegrationReader,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during integration ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own the integrations defined in the Dag.")
	}

	// If a workflow with the same name already exists for the user, we will treat this as an
	// update to the workflow instead of creation.
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
		dagSummary.Dag.WorkflowId = collidingWorkflow.Id
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

	return &registerWorkflowArgs{
		AqContext:                aqContext,
		dbWorkflowDag:            dagSummary.Dag,
		operatorIdToFileContents: dagSummary.FileContentsByOperatorUUID,
		isUpdate:                 isUpdate,
	}, http.StatusOK, nil
}

func (h *RegisterWorkflowHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*registerWorkflowArgs)

	emptyResp := registerWorkflowResponse{}

	if _, err := operator_utils.UploadOperatorFiles(ctx, args.dbWorkflowDag, args.operatorIdToFileContents); err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to create workflow.")
	}

	txn, err := h.Database.BeginTx(ctx)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to create workflow.")
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	workflowId, err := utils.WriteWorkflowDagToDatabase(
		ctx,
		args.dbWorkflowDag,
		h.WorkflowReader,
		h.WorkflowWriter,
		h.WorkflowDagWriter,
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

	args.dbWorkflowDag.Metadata.Id = workflowId

	// workflowDag, err := dag_utils.NewWorkflowDag(
	// 	ctx,
	// 	args.dbWorkflowDag,
	// 	h.WorkflowDagResultWriter,
	// 	h.OperatorResultWriter,
	// 	h.ArtifactResultWriter,
	// 	h.WorkflowReader,
	// 	h.NotificationWriter,
	// 	h.UserReader,
	// 	h.JobManager,
	// 	h.Vault,
	// 	h.StorageConfig,
	// 	h.Database,
	// )
	// if err != nil {
	// 	return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Error creating dag object.")
	// }

	eng, err := engine.NewAqEngine(
		h.Database,
		h.GithubManager,
		h.Vault,
		h.JobManager,
		h.StorageConfig,
		engine.AqueductTimeConfig{
			OperatorPollInterval: previewPollIntervalMillisec,
			ExecTimeout:          engine.DefaultExecutionTimeout,
			CleanupTimeout:       engine.DefaultCleanupTimeout,
		},
		true, /* shouldPersistResults */
		h.WorkflowDagResultWriter,
		h.OperatorResultWriter,
		h.ArtifactResultWriter,
		h.NotificationWriter,
		h.WorkflowReader,
		h.UserReader,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Error creating engine.")
	}

	if args.isUpdate {
		// If we're updating an existing workflow, first update the metadata.
		_, _, err = (&EditWorkflowHandler{
			Database:       txn,
			WorkflowReader: h.WorkflowReader,
			WorkflowWriter: h.WorkflowWriter,
			JobManager:     h.JobManager,
		}).Perform(
			ctx,
			&editWorkflowArgs{
				workflowId:          workflowId,
				workflowName:        args.dbWorkflowDag.Metadata.Name,
				workflowDescription: args.dbWorkflowDag.Metadata.Description,
				schedule:            &args.dbWorkflowDag.Metadata.Schedule,
			},
		)
		if err != nil {
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to update existing workflow.")
		}
	} else {
		// We should create cron jobs for newly created, non-manually triggered workflows.
		if string(args.dbWorkflowDag.Metadata.Schedule.CronSchedule) != "" {

			err = eng.ScheduleWorkflow(
				ctx,
				args.dbWorkflowDag,
				shared_utils.AppendPrefix(args.dbWorkflowDag.Metadata.Id.String()),
				string(args.dbWorkflowDag.Metadata.Schedule.CronSchedule),
			)

			if err != nil {
				return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to create workflow.")
			}
		}
	}

	if err := txn.Commit(ctx); err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to create workflow.")
	}

	_, err = eng.ExecuteWorkflow(ctx, args.dbWorkflowDag)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to trigger workflow run.")
	}

	if !args.isUpdate {
		// If this workflow is newly created, automatically add the user to the workflow's
		// watchers list.
		watchWorkflowArgs := &watchWorkflowArgs{
			AqContext:  args.AqContext,
			workflowId: workflowId,
		}

		_, _, err = (&WatchWorkflowHandler{
			Database:              h.Database,
			WorkflowReader:        h.WorkflowReader,
			WorkflowWatcherWriter: h.WorkflowWatcherWriter,
		}).Perform(ctx, watchWorkflowArgs)
		if err != nil {
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to add user who created the workflow to watch.")
		}
	}

	return registerWorkflowResponse{Id: workflowId}, http.StatusOK, nil
}

// CreateWorkflowCronJob creates a k8s cron job
// that will run the workflow on the specified schedule.
func CreateWorkflowCronJob(
	ctx context.Context,
	workflow *workflow.Workflow,
	dbConfig *database.DatabaseConfig,
	vaultObject vault.Vault,
	jobManager job.JobManager,
	githubManager github.Manager,
) error {
	workflowId := workflow.Id.String()
	name := shared_utils.AppendPrefix(workflowId)
	period := string(workflow.Schedule.CronSchedule)

	spec := job.NewWorkflowSpec(
		workflow.Name,
		workflowId,
		dbConfig,
		vaultObject.Config(),
		jobManager.Config(),
		githubManager.Config(),
		nil, /* parameters */
	)

	err := jobManager.DeployCronJob(
		ctx,
		name,
		period,
		spec,
	)
	if err != nil {
		return errors.Wrap(err, "unable to deploy workflow cron job")
	}
	return nil
}
