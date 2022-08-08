package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/workflow/engine"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// The `DeleteWorkflowHandler` does a best effort at deleting a workflow and its dependencies, such as
// k8s resources, Postgres state, and output tables in the user's data warehouse.
type deleteWorkflowArgs struct {
	*aq_context.AqContext
	workflowId uuid.UUID
}

type deleteWorkflowResponse struct{}

type DeleteWorkflowHandler struct {
	PostHandler

	Database       database.Database
	WorkflowReader workflow.Reader
	Engine         engine.Engine
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

	return &deleteWorkflowArgs{
		AqContext:  aqContext,
		workflowId: workflowId,
	}, http.StatusOK, nil
}

func (h *DeleteWorkflowHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*deleteWorkflowArgs)

	emptyResp := deleteWorkflowResponse{}

	err := h.Engine.DeleteWorkflow(ctx, args.workflowId)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to delete workflow.")
	}

	return emptyResp, http.StatusOK, nil
}
