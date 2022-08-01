package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/request"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/user"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_result"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/vault"
	dag_utils "github.com/aqueducthq/aqueduct/lib/workflow/dag"
	"github.com/aqueducthq/aqueduct/lib/workflow/engine"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/github"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type RefreshWorkflowArgs struct {
	WorkflowId uuid.UUID
	Parameters map[string]string
}

// Refresh workflow creates a new workflow version by
// triggering running a workflow run.
type RefreshWorkflowHandler struct {
	PostHandler

	Database      database.Database
	JobManager    job.JobManager
	GithubManager github.Manager
	Vault         vault.Vault
	StorageConfig *shared.StorageConfig

	OperatorReader        operator.Reader
	ArtifactReader        artifact.Reader
	WorkflowDagEdgeReader workflow_dag_edge.Reader
	WorkflowReader        workflow.Reader
	WorkflowDagReader     workflow_dag.Reader
	UserReader            user.Reader

	ArtifactResultWriter    artifact_result.Writer
	OperatorResultWriter    operator_result.Writer
	NotificationWriter      notification.Writer
	WorkflowDagResultWriter workflow_dag_result.Writer
}

func (*RefreshWorkflowHandler) Name() string {
	return "RefreshWorkflow"
}

func (h *RefreshWorkflowHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	workflowIdStr := chi.URLParam(r, routes.WorkflowIdUrlParam)
	if workflowIdStr == "" {
		return nil, http.StatusBadRequest, errors.New("no workflow id was specified")
	}

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

	parameters, err := request.ExtractParamsfromRequest(r)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The user-defined parameters could not be extracted in current format.")
	}

	return &RefreshWorkflowArgs{
		WorkflowId: workflowId,
		Parameters: parameters,
	}, http.StatusOK, nil
}

func (h *RefreshWorkflowHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*RefreshWorkflowArgs)

	dag, err := utils.ReadLatestWorkflowDagFromDatabase(
		ctx,
		args.WorkflowId,
		h.WorkflowReader,
		h.WorkflowDagReader,
		h.OperatorReader,
		h.ArtifactReader,
		h.WorkflowDagEdgeReader,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Error reading dag object from database.")
	}

	if args.Parameters != nil {
		for name, newVal := range args.Parameters {
			op := dag.GetOperatorByName(name)
			if op == nil {
				continue
			}
			if !op.Spec.IsParam() {
				return nil, http.StatusInternalServerError, errors.Newf("Cannot set parameters on a non-parameter operator %s", name)
			}
			dag.Operators[op.Id].Spec.Param().Val = newVal
		}
	}

	workflowDag, err := dag_utils.NewWorkflowDag(
		ctx,
		dag,
		h.WorkflowDagResultWriter,
		h.OperatorResultWriter,
		h.ArtifactResultWriter,
		h.WorkflowReader,
		h.NotificationWriter,
		h.UserReader,
		h.JobManager,
		h.Vault,
		h.StorageConfig,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Error creating dag object.")
	}

	eng, err := engine.NewAqEngine(
		workflowDag,
		h.Database,
		h.GithubManager,
		h.Vault,
		h.JobManager,
		engine.AqueductTimeConfig{
			OperatorPollInterval: previewPollIntervalMillisec,
			ExecTimeout:          engine.DefaultExecutionTimeout,
			CleanupTimeout:       engine.DefaultCleanupTimeout,
		},
		true,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Error creating Aqueduct Engine.")
	}

	_, err = eng.ExecuteWorkflow(ctx, workflowDag)

	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Error executing the workflow.")
	}

	return struct{}{}, http.StatusOK, nil
}
