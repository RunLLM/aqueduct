package v2

import (
	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/cmd/server/request/parser"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/google/uuid"
	"net/http"
)

// This file should map directly to
// src/ui/common/src/handlers/v2/NodesOperatorsGet.ts
//
// Route: /api/v2/workflow/{workflowID}/dag/{dagID}/nodes/operators
// Method: GET
// Params:
//	`workflowID`: ID for `workflow` object
//  `dagID`: ID for `workflow_dag` object
// Request:
//	Headers:
//		`api-key`: user's API Key
// Parameters:
//		`integrationID`:
//            Optional field that filters out the given results to only those that belong to the given integration.
// Response:
//	Body:
//    A list of `response.operators` that belong to the given dag.

type nodesOperatorsGetArgs struct {
	*aq_context.AqContext
	integrationID *uuid.UUID
	workflowID    uuid.UUID
	dagID         uuid.UUID
}

type NodeOperatorsGetHandler struct {
	handler.GetHandler

	Database        database.Database
	IntegrationRepo repos.Integration
	OperatorRepo    repos.Operator
}

func (*NodeOperatorsGetHandler) Name() string {
	return "NodesOperatorsGet"
}

func (h *NodeOperatorsGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
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

	integrationID, err := (parser.IntegrationIDParser{}).Parse(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return &nodesOperatorsGetArgs{
		AqContext: aqContext,

		integrationID: integrationID,
		workflowID:    workflowID,
		dagID:         dagID,
	}, http.StatusOK, nil
}

func (h *NodeOperatorsGetHandler) Get(r *http.Request, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*nodesOperatorsGetArgs)

}
