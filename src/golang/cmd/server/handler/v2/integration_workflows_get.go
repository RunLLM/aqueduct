package v2

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/cmd/server/request/parser"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/google/uuid"
)

// This file should map directly to
// src/ui/common/src/handlers/v2/IntegrationWorkflowsGet.ts
//
// Route: /v2/integration/{integrationID}/workflows
// Method: GET
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//   	A list of `response.WorkflowAndDagID` that use the given integration.

type integrationWorkflowsGetArgs struct {
	*aq_context.AqContext
	integrationID uuid.UUID
}

type IntegrationWorkflowsGetHandler struct {
	handler.GetHandler

	Database        database.Database
	IntegrationRepo repos.Integration
	WorkflowRepo    repos.Workflow
	DAGRepo         repos.DAG
	DAGResultRepo   repos.DAGResult
	OperatorRepo    repos.Operator
}

func (*IntegrationWorkflowsGetHandler) Name() string {
	return "IntegrationWorkflowsGet"
}

func (h *IntegrationWorkflowsGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	integrationID, err := (parser.IntegrationIDParser{}).Parse(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return &integrationWorkflowsGetArgs{
		AqContext:     aqContext,
		integrationID: *integrationID,
	}, http.StatusOK, nil
}

func (h *IntegrationWorkflowsGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*integrationWorkflowsGetArgs)

	integration, err := h.IntegrationRepo.Get(ctx, args.integrationID, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "Unable to find integration %s", args.integrationID)
	}

	workflowAndDagIDs, err := fetchWorkflowAndDagIDsForIntegration(
		ctx, args.OrgID, integration, h.IntegrationRepo, h.WorkflowRepo, h.OperatorRepo, h.DAGRepo, h.DAGResultRepo, h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "Unable to find workflows for integration %s", args.integrationID)
	}
	return workflowAndDagIDs, http.StatusOK, nil
}
