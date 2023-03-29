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
// src/ui/common/src/handlers/DagResultGet.tsx
//
// Route: /v2/workflow/{workflowId}
// Method: GET
// Params:
//	`workflowId`: ID for `workflow` object
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		serialized `workflow.Response`

type dagResultGetArgs struct {
	*aq_context.AqContext
	workflowID  uuid.UUID
	dagResultID uuid.UUID
}

type DAGResultGetHandler struct {
	handler.GetHandler

	Database database.Database

	WorkflowRepo  repos.Workflow
	DAGResultRepo repos.DAGResult
}

func (*DAGResultGetHandler) Name() string {
	return "DAGResultGet"
}

func (h *DAGResultGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	workflowIDStr := chi.URLParam(r, routes.WorkflowIDParam)
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed workflow ID.")
	}

	dagResultIDStr := chi.URLParam(r, routes.DAGResultIDParam)
	dagResultID, err := uuid.Parse(dagResultIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed DAG ID.")
	}

	return &dagResultGetArgs{
		AqContext:   aqContext,
		workflowID:  workflowID,
		dagResultID: dagResultID,
	}, http.StatusOK, nil
}

func (h *DAGResultGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*dagResultGetArgs)

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

	dbDAGResult, err := h.DAGResultRepo.Get(ctx, args.dagResultID, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error reading workflow.")
	}

	return response.NewDAGResultFromDBObject(dbDAGResult), http.StatusOK, nil
}
