package v2

import (
	"context"
	"fmt"
	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/cmd/server/request/parser"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/google/uuid"
	"net/http"
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
//   	A list of workflow IDs that use the given integration.

type integrationWorkflowsGetArgs struct {
	*aq_context.AqContext
	integrationID uuid.UUID
}

type IntegrationWorkflowsGetHandler struct {
	handler.GetHandler

	Database        database.Database
	IntegrationRepo repos.Integration
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
		integrationID: integrationID,
	}, http.StatusOK, nil
}

func (h *IntegrationWorkflowsGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*integrationWorkflowsGetArgs)

	integration, err := h.IntegrationRepo.Get(ctx, args.integrationID, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, fmt.Sprintf("Unable to find integration %s", args.integrationID))
	}

	workflowIDs, err := fetchWorkflowIDsForIntegration(
		ctx, args.OrgID, integration, h.IntegrationRepo, h.OperatorRepo, h.Database,
	)
	return workflowIDs, http.StatusOK, nil
}
