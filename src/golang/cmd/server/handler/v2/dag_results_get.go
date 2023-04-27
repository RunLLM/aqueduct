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
// src/ui/common/src/handlers/v2/DagResultsGet.tsx
//
// Route: /v2/workflow/{workflowId}/results
// Method: GET
// Params:
//	`workflowId`: ID for `workflow` object
// Request:
//	Headers:
//		`api-key`: user's API Key
//  Parameters:
//		`order_by`:
//			Optional single field that the query should be ordered. Requires the table prefix.
//		`order_descending`:
//			Optional boolean specifying whether order_by should be ascending or descending.
//		`limit`:
//			Optional limit on the number of storage migrations returned. Defaults to all of them.
// Response:
//	Body:
//		serialized `[]response.DAGResult`

type dagResultsGetArgs struct {
	*aq_context.AqContext
	workflowID uuid.UUID

	// A nil value means that the order is not set.
	orderBy string
	// Default is descending (true).
	orderDescending bool
	// A negative value for limit (eg. -1) means that the limit is not set.
	limit int
}

type DAGResultsGetHandler struct {
	handler.GetHandler

	Database database.Database

	DAGResultRepo repos.DAGResult
}

func (*DAGResultsGetHandler) Name() string {
	return "DAGResultsGet"
}

func (h *DAGResultsGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	workflowID, err := (parser.WorkflowIDParser{}).Parse(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	limit, err := (parser.LimitQueryParser{}).Parse(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	orderBy, err := (parser.OrderByQueryParser{}).Parse(r, models.AllDAGResultCols())
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	descending, err := (parser.OrderDescendingQueryParser{}).Parse(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return &dagResultsGetArgs{
		AqContext:       aqContext,
		workflowID:      workflowID,
		orderBy:         orderBy,
		orderDescending: descending,
		limit:           limit,
	}, http.StatusOK, nil
}

func (h *DAGResultsGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*dagResultsGetArgs)

	dbDAGResults, err := h.DAGResultRepo.GetByWorkflow(ctx, args.workflowID, args.orderBy, args.limit, args.orderDescending, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error reading dag results.")
	}

	return slices.Map(dbDAGResults, func(dbResult models.DAGResult) response.DAGResult {
		return *response.NewDAGResultFromDBObject(&dbResult)
	}), http.StatusOK, nil
}
