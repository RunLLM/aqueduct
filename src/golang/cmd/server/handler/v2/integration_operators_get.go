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

type resourceOperatorsGetArgs struct {
	*aq_context.AqContext
	resourceID uuid.UUID
}

type ResourceOperatorsGetHandler struct {
	handler.GetHandler

	Database           database.Database
	ResourceRepo       repos.Resource
	OperatorRepo       repos.Operator
	OperatorResultRepo repos.OperatorResult
}

func (*ResourceOperatorsGetHandler) Name() string {
	return "IntegrationOperatorsGet"
}

func (h *ResourceOperatorsGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	resourceID, err := (parser.ResourceIDParser{}).Parse(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return &resourceOperatorsGetArgs{
		AqContext: aqContext,

		resourceID: *resourceID,
	}, http.StatusOK, nil
}

func (h *ResourceOperatorsGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*resourceOperatorsGetArgs)

	resource, err := h.ResourceRepo.Get(ctx, args.resourceID, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "Unable to find resource %s", args.resourceID)
	}

	operators, err := operator.GetOperatorsOnResource(
		ctx,
		args.OrgID,
		resource,
		h.ResourceRepo,
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
