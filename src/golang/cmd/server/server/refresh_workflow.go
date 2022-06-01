package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/utils"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/github"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type refreshWorkflowArgs struct {
	workflowId uuid.UUID
}

// Refresh workflow creates a new workflow version by
// triggering running a workflow run.
type RefreshWorkflowHandler struct {
	PostHandler

	Database       database.Database
	JobManager     job.JobManager
	GithubManager  github.Manager
	Vault          vault.Vault
	WorkflowReader workflow.Reader
}

func (*RefreshWorkflowHandler) Name() string {
	return "RefreshWorkflow"
}

func (h *RefreshWorkflowHandler) Prepare(r *http.Request) (interface{}, int, error) {
	common, statusCode, err := ParseCommonArgs(r)
	if err != nil {
		return nil, statusCode, err
	}

	workflowIdStr := chi.URLParam(r, utils.WorkflowIdUrlParam)
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
		common.OrganizationId,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during workflow ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this workflow.")
	}

	return &refreshWorkflowArgs{
		workflowId: workflowId,
	}, http.StatusOK, nil
}

func generateWorkflowJobName() string {
	return fmt.Sprintf("workflow-adhoc-%s", uuid.New().String())
}

func (h *RefreshWorkflowHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*refreshWorkflowArgs)

	workflowObject, err := h.WorkflowReader.GetWorkflow(
		ctx,
		args.workflowId,
		h.Database,
	)
	if err != nil {
		if err == database.ErrNoRows {
			return nil, http.StatusBadRequest, errors.New("Unable to find workflow.")
		}
		return nil, http.StatusInternalServerError, errors.New("Unable to find workflow.")
	}

	jobName := generateWorkflowJobName()

	jobSpec := job.NewWorkflowSpec(
		workflowObject.Name,
		workflowObject.Id.String(),
		h.Database.Config(),
		h.Vault.Config(),
		h.JobManager.Config(),
		h.GithubManager.Config(),
	)

	err = h.JobManager.Launch(
		ctx,
		jobName,
		jobSpec,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to trigger this workflow.")
	}
	return struct{}{}, http.StatusOK, nil
}
