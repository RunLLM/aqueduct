package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/request"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	exec_env "github.com/aqueducthq/aqueduct/lib/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/job"
	shared_utils "github.com/aqueducthq/aqueduct/lib/lib_utils"
	mdl_utils "github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/workflow"
	dag_utils "github.com/aqueducthq/aqueduct/lib/workflow/dag"
	operator_utils "github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/github"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
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
	Engine        engine.Engine

	ArtifactRepo             repos.Artifact
	DAGRepo                  repos.DAG
	DAGEdgeRepo              repos.DAGEdge
	ExecutionEnvironmentRepo repos.ExecutionEnvironment
	IntegrationRepo          repos.Integration
	OperatorRepo             repos.Operator
	WatcherRepo              repos.Watcher
	WorkflowRepo             repos.Workflow
}

type registerWorkflowArgs struct {
	*aq_context.AqContext
	dagSummary *request.DagSummary

	// Whether this is a registering a new workflow or updating an existing one.
	isUpdate bool
	runNow   bool
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

	runNowStr := r.Header.Get(routes.RunNowHeader)
	runNow := true
	if runNowStr == "False" {
		runNow = false
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
		h.IntegrationRepo,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during integration ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own the integrations defined in the Dag.")
	}

	isUpdate := true
	// If a workflow with the same name already exists for the user, we will treat this as an
	// update to the workflow instead of creation.
	collidingWorkflow, err := h.WorkflowRepo.GetByOwnerAndName(
		r.Context(),
		dagSummary.Dag.Metadata.UserID,
		dagSummary.Dag.Metadata.Name,
		h.Database,
	)
	if err != nil {
		if err != database.ErrNoRows {
			return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when checking for existing workflows.")
		}
		// A colliding workflow does not exist, so this is not an update
		isUpdate = false
	}

	if isUpdate {
		// Since the libraries we call use the workflow id to tell whether a workflow already exists.
		dagSummary.Dag.WorkflowID = collidingWorkflow.ID
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
		AqContext:  aqContext,
		dagSummary: dagSummary,
		isUpdate:   isUpdate,
		runNow:     runNow,
	}, http.StatusOK, nil
}

func (h *RegisterWorkflowHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*registerWorkflowArgs)
	dbWorkflowDag := args.dagSummary.Dag
	fileContentsByOperatorID := args.dagSummary.FileContentsByOperatorUUID

	emptyResp := registerWorkflowResponse{}

	if _, err := operator_utils.UploadOperatorFiles(
		ctx,
		dbWorkflowDag,
		fileContentsByOperatorID,
	); err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to create workflow.")
	}

	execEnvByOpId, status, err := setupExecEnv(
		ctx,
		args.ID,
		args.dagSummary,
		h.IntegrationRepo,
		h.ExecutionEnvironmentRepo,
		h.Database,
	)
	if err != nil {
		return emptyResp, status, err
	}

	for opId, op := range args.dagSummary.Dag.Operators {
		if env, ok := execEnvByOpId[opId]; ok {
			// Note: this is the canotical way to assign a struct field of a map
			// https://stackoverflow.com/questions/42605337/cannot-assign-to-struct-field-in-a-map
			op.ExecutionEnvironmentID = mdl_utils.NullUUID{UUID: env.ID, IsNull: false}
			dbWorkflowDag.Operators[opId] = op
		}
	}

	txn, err := h.Database.BeginTx(ctx)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to create workflow.")
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	// Schedule validation needs to happen inside the `txn` to prevent
	// concurrent requests from forming a cycle among cascading workflows
	validateScheduleCode, err := workflow.ValidateSchedule(
		ctx,
		args.isUpdate,
		dbWorkflowDag.WorkflowID,
		dbWorkflowDag.Metadata.Schedule,
		dbWorkflowDag.EngineConfig.Type,
		h.ArtifactRepo,
		h.DAGRepo,
		h.DAGEdgeRepo,
		h.OperatorRepo,
		h.WorkflowRepo,
		h.Database,
	)
	if err != nil {
		return emptyResp, validateScheduleCode, err
	}

	workflowId, err := utils.WriteDAGToDatabase(
		ctx,
		dbWorkflowDag,
		h.WorkflowRepo,
		h.DAGRepo,
		h.OperatorRepo,
		h.DAGEdgeRepo,
		h.ArtifactRepo,
		txn,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to create workflow.")
	}

	args.dagSummary.Dag.Metadata.ID = workflowId

	if args.isUpdate {
		// If we're updating an existing workflow, first update the metadata.
		err := h.Engine.EditWorkflow(
			ctx,
			txn,
			workflowId,
			dbWorkflowDag.Metadata.Name,
			dbWorkflowDag.Metadata.Description,
			&dbWorkflowDag.Metadata.Schedule,
			&dbWorkflowDag.Metadata.RetentionPolicy,
			&dbWorkflowDag.Metadata.NotificationSettings,
		)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to update workflow.")
		}

	} else {
		// We should create cron jobs for newly created, non-manually triggered workflows.
		if string(dbWorkflowDag.Metadata.Schedule.CronSchedule) != "" {

			err = h.Engine.ScheduleWorkflow(
				ctx,
				workflowId,
				shared_utils.AppendPrefix(dbWorkflowDag.Metadata.ID.String()),
				string(dbWorkflowDag.Metadata.Schedule.CronSchedule),
			)

			if err != nil {
				return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to create workflow.")
			}
		}
	}

	if err := txn.Commit(ctx); err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to create workflow.")
	}

	timeConfig := &engine.AqueductTimeConfig{
		OperatorPollInterval: engine.DefaultPollIntervalMillisec,
		ExecTimeout:          engine.DefaultExecutionTimeout,
		CleanupTimeout:       engine.DefaultCleanupTimeout,
	}

	if args.runNow {
		_, err = h.Engine.TriggerWorkflow(
			ctx,
			workflowId,
			shared_utils.AppendPrefix(dbWorkflowDag.Metadata.ID.String()),
			timeConfig,
			nil, /* parameters */
		)
		if err != nil {
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to trigger workflow.")
		}
	}

	if !args.isUpdate {
		// If this workflow is newly created, automatically add the user to the workflow's
		// watchers list.
		watchWorkflowArgs := &watchWorkflowArgs{
			AqContext:  args.AqContext,
			workflowId: workflowId,
		}

		_, _, err = (&WatchWorkflowHandler{
			Database: h.Database,

			WatcherRepo:  h.WatcherRepo,
			WorkflowRepo: h.WorkflowRepo,
		}).Perform(ctx, watchWorkflowArgs)
		if err != nil {
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to add user who created the workflow to watch.")
		}
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
			h.ExecutionEnvironmentRepo,
			db,
		)
		if err != nil {
			log.Errorf("%v", err)
		}
	}()

	return registerWorkflowResponse{Id: workflowId}, http.StatusOK, nil
}
