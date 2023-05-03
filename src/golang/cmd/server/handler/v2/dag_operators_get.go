package v2

import (
	"context"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/functional/slices"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/aqueducthq/aqueduct/lib/response"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/cmd/server/request/parser"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/google/uuid"
)

// This file should map directly to
// src/ui/common/src/handlers/v2/DagOperatorsGet.ts
//
// Route: /api/v2/workflow/{workflowId}/dag/{dagID}/nodes/operators
// Method: GET
// Params:
//	`workflowId`: ID for `workflow` object
//  `dagID`: ID for `workflow_dag` object
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//    A list of `response.operators` on the dag result.

type dagOperatorsArgs struct {
	*aq_context.AqContext
	workflowID uuid.UUID
	dagID      uuid.UUID
}

type DagOperatorsGetHandler struct {
	handler.GetHandler

	Database     database.Database
	WorkflowRepo repos.Workflow
	OperatorRepo repos.Operator
}

func (*DagOperatorsGetHandler) Name() string {
	return "DagResultOperators"
}

func (h *DagOperatorsGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
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

	return &dagOperatorsArgs{
		AqContext:  aqContext,
		workflowID: workflowID,
		dagID:      dagID,
	}, http.StatusOK, nil
}

func (h *DagOperatorsGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*dagOperatorsArgs)

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

	operatorNodes, err := h.OperatorRepo.GetNodesByDAG(ctx, args.dagID, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during operator retrieval.")
	}

	return slices.Map(operatorNodes, func(node views.OperatorNode) *response.Operator {
		return response.NewOperatorFromDBObject(&node)
	}), http.StatusOK, nil
}
