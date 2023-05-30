package v2

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/cmd/server/request/parser"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/functional/slices"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/response"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// This file should map directly to
// src/ui/common/src/handlers/v2/DagsGet.tsx
//
// Route: /v2/workflow/{workflowId}/dags
// Method: GET
// Params:
//	`workflowId`: ID for `workflow` object
// Request:
//	Headers:
//		`api-key`: user's API Key
//
// Response:
//	Body:
//		serialized `[]response.DAGResult`

type dagsGetArgs struct {
	*aq_context.AqContext
	workflowID uuid.UUID
}

type DAGsGetHandler struct {
	handler.GetHandler

	Database database.Database

	WorkflowRepo repos.Workflow
	DAGRepo      repos.DAG
}

func (*DAGsGetHandler) Name() string {
	return "DAGsGet"
}

func (h *DAGsGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	workflowID, err := (parser.WorkflowIDParser{}).Parse(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return &dagsGetArgs{
		AqContext:  aqContext,
		workflowID: workflowID,
	}, http.StatusOK, nil
}

func (h *DAGsGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*dagsGetArgs)

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

	dbDAGs, err := h.DAGRepo.GetByWorkflow(ctx, args.workflowID, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error reading dag results.")
	}

	return slices.Map(dbDAGs, func(dbDAG models.DAG) response.DAG {
		return *response.NewDAGFromDBObject(&dbDAG)
	}), http.StatusOK, nil
}
