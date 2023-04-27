package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/response"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type unwatchWorkflowArgs struct {
	*aq_context.AqContext
	workflowId uuid.UUID
}

// Route: /workflow/{workflowId}/unwatch
// Method: POST
// Params: workflowId
// Request
//
//	Headers:
//		`api-key`: user's API Key
//
// Response: None
type UnwatchWorkflowHandler struct {
	PostHandler

	Database database.Database

	WatcherRepo repos.Watcher
}

func (*UnwatchWorkflowHandler) Name() string {
	return "UnwatchWorkflow"
}

func (h *UnwatchWorkflowHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	workflowIdStr := chi.URLParam(r, routes.WorkflowIdUrlParam)
	workflowId, err := uuid.Parse(workflowIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed workflow ID.")
	}

	return &unwatchWorkflowArgs{
		AqContext:  aqContext,
		workflowId: workflowId,
	}, http.StatusOK, nil
}

func (h *UnwatchWorkflowHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*unwatchWorkflowArgs)

	response := response.EmptyResponse{}

	err := h.WatcherRepo.Delete(ctx, args.workflowId, args.ID, h.Database)
	if err != nil {
		return response, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error while updating the database.")
	}

	return response, http.StatusOK, nil
}
