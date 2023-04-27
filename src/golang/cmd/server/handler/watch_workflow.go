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

type watchWorkflowArgs struct {
	*aq_context.AqContext
	workflowId uuid.UUID
}

// Route: /workflow/{workflowId}/watch
// Method: POST
// Params: workflowId
// Request
//
//	Headers:
//		`api-key`: user's API Key
//
// Response: None
type WatchWorkflowHandler struct {
	PostHandler

	Database database.Database

	WatcherRepo repos.Watcher
}

func (*WatchWorkflowHandler) Name() string {
	return "WatchWorkflow"
}

func (h *WatchWorkflowHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	workflowIDStr := chi.URLParam(r, routes.WorkflowIdUrlParam)
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed workflow ID.")
	}

	return &watchWorkflowArgs{
		AqContext:  aqContext,
		workflowId: workflowID,
	}, http.StatusOK, nil
}

func (h *WatchWorkflowHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*watchWorkflowArgs)

	response := response.EmptyResponse{}

	_, err := h.WatcherRepo.Create(ctx, args.workflowId, args.ID, h.Database)
	if err != nil {
		return response, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error while updating the database.")
	}

	return response, http.StatusOK, nil
}
