package v2

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/cmd/server/request/parser"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/response"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// This file should map directly to
// src/ui/common/src/handlers/DagGet.tsx
//
// Route: /v2/workflow/{workflowID}/dag/{dagID}
// Method: GET
// Params:
//	`workflowID`: ID for `workflow` object
//  `dagID`: ID for `workflow_dag` object
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		serialized `workflow.Response`

type dagGetArgs struct {
	*aq_context.AqContext
	workflowID uuid.UUID
	dagID      uuid.UUID
}

type DAGGetHandler struct {
	handler.GetHandler

	Database database.Database

	WorkflowRepo repos.Workflow
	DAGRepo      repos.DAG
}

func (*DAGGetHandler) Name() string {
	return "DAGGet"
}

func (h *DAGGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	workflowID, err := (parser.WorkflowIDParser{}).Parse(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	dagID, err := (parser.DagIDParser{}).Parse(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return &dagGetArgs{
		AqContext:  aqContext,
		workflowID: workflowID,
		dagID:      dagID,
	}, http.StatusOK, nil
}

func (h *DAGGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*dagGetArgs)

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

	dbDAG, err := h.DAGRepo.Get(ctx, args.dagID, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error reading workflow.")
	}

	return response.NewDAGFromDBObject(dbDAG), http.StatusOK, nil
}
