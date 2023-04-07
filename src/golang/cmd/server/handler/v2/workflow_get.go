package v2

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/response"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// This file should map directly to
// src/ui/common/src/handlers/WorkflowGet.tsx
//
// Route: /v2/workflow/{workflowID}
// Method: GET
// Params:
//	`workflowId`: ID for `workflow` object
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		serialized `workflow.Response`

type workflowGetArgs struct {
	*aq_context.AqContext
	workflowID uuid.UUID
}

type WorkflowGetHandler struct {
	handler.GetHandler

	Database database.Database

	WorkflowRepo repos.Workflow
}

func (*WorkflowGetHandler) Name() string {
	return "WorkflowGet"
}

func (h *WorkflowGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	workflowIDStr := chi.URLParam(r, routes.WorkflowIDParam)
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed workflow ID.")
	}

	return &workflowGetArgs{
		AqContext:  aqContext,
		workflowID: workflowID,
	}, http.StatusOK, nil
}

func (h *WorkflowGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*workflowGetArgs)

	ok, err := h.WorkflowRepo.ValidateOrg(
		ctx,
		args.workflowID,
		args.OrgID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during workflow ownership validation.")
	}

	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this workflow.")
	}

	dbWorkflow, err := h.WorkflowRepo.Get(ctx, args.workflowID, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error reading workflow.")
	}

	return response.NewWorkflowFromDBObject(dbWorkflow), http.StatusOK, nil
}
