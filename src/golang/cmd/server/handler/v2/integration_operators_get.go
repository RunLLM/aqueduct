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
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/response"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// This file should map directly to
// src/ui/common/src/handlers/v2/IntegrationOperatorsGet.ts
//
// Route: /api/v2/integration/{integrationID}/nodes/operators
// Method: GET
// Params:
// 	`integrationID`: ID for `integration` object
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//    A list of `response.operators` that use the given integration.

type integrationOperatorsGetArgs struct {
	*aq_context.AqContext
	integrationID uuid.UUID
}

type IntegrationOperatorsGetHandler struct {
	handler.GetHandler

	Database           database.Database
	IntegrationRepo    repos.Integration
	OperatorRepo       repos.Operator
	OperatorResultRepo repos.OperatorResult
}

func (*IntegrationOperatorsGetHandler) Name() string {
	return "IntegrationOperatorsGet"
}

func (h *IntegrationOperatorsGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	integrationID, err := (parser.IntegrationIDParser{}).Parse(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return &integrationOperatorsGetArgs{
		AqContext: aqContext,

		integrationID: *integrationID,
	}, http.StatusOK, nil
}

func (h *IntegrationOperatorsGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*integrationOperatorsGetArgs)

	integration, err := h.IntegrationRepo.Get(ctx, args.integrationID, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "Unable to find integration %s", args.integrationID)
	}

	operators, err := operator.GetOperatorsOnIntegration(
		ctx,
		args.OrgID,
		integration,
		h.IntegrationRepo,
		h.OperatorRepo,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve operators.")
	}
	if len(operators) == 0 {
		return []*response.Operator{}, http.StatusOK, nil
	}

	operatorIDs := slices.Map(operators, func(op models.Operator) uuid.UUID {
		return op.ID
	})
	operatorNodes, err := h.OperatorRepo.GetNodeBatch(ctx, operatorIDs, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve operator nodes.")
	}

	return slices.Map(operatorNodes, func(node views.OperatorNode) *response.Operator {
		return response.NewOperatorFromDBObject(&node)
	}), http.StatusOK, nil
}
