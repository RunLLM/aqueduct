package handler

import (
	"context"
	"fmt"
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

	Database                database.Database
	JobManager              job.JobManager
	GithubManager           github.Manager
	Vault                   vault.Vault
	StorageConfig           *shared.StorageConfig
	OperatorReader          operator.Reader
	ArtifactReader          artifact.Reader
	ArtifactResultWriter    artifact_result.Writer
	OperatorResultWriter    operator_result.Writer
	NotificationWriter      notification.Writer
	WorkflowDagEdgeReader   workflow_dag_edge.Reader
	WorkflowDagResultWriter workflow_dag_result.Writer
	WorkflowReader          workflow.Reader
	WorkflowDagReader       workflow_dag.Reader
	UserReader              user.Reader
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

func generateWorkflowJobName() string {
	return fmt.Sprintf("workflow-adhoc-%s", uuid.New().String())
}

func (h *RefreshWorkflowHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*RefreshWorkflowArgs)
	
	//extract the dag based on workflow ID provided.
	dag, err := h.WorkflowDagReader.GetLatestWorkflowDag(ctx, args.WorkflowId, h.Database)

	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Error creating dag object.")
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

	_ , err = eng.ExecuteWorkflow(ctx,workflowDag)
	// jobName := generateWorkflowJobName()

	// jobSpec := job.NewWorkflowSpec(
	// 	workflowObject.Name,
	// 	workflowObject.Id.String(),
	// 	h.Database.Config(),
	// 	h.Vault.Config(),
	// 	h.JobManager.Config(),
	// 	h.GithubManager.Config(),
	// 	args.Parameters,
	// )

	// err = h.JobManager.Launch(
	// 	ctx,
	// 	jobName,
	// 	jobSpec,
	// )
	if err != nil {
	return nil, http.StatusInternalServerError, errors.Wrap(err, "Error executing the workflow.")
	}
	return struct{}{}, http.StatusOK, nil
}
