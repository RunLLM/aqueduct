package v2

import (
	"context"
	"net/http"
	"strconv"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
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
//		`order_by`:
//			Optional single field that the query should be ordered. Requires the table prefix.
//		`limit`:
//			Optional limit on the number of storage migrations returned. Defaults to all of them.
// Response:
//	Body:
//		serialized `[]response.DAGResult`

type dagResultsGetArgs struct {
	*aq_context.AqContext
	workflowID uuid.UUID
	
	// A nil value means that the order is not set.
	orderBy         string
	// A negative value for limit (eg. -1) means that the limit is not set.
	limit int
}

type DAGResultsGetHandler struct {
	handler.GetHandler

	Database database.Database

	WorkflowRepo  repos.Workflow
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

	limit := -1
	if limitVal := r.Header.Get(routes.DagResultGetLimitHeader); len(limitVal) > 0 {
		limit, err = strconv.Atoi(limitVal)
		if err != nil {
			return nil, http.StatusBadRequest, errors.Wrap(err, "Invalid limit header.")
		}
	}

	var orderBy string
	if orderByVal := r.Header.Get(routes.DagResultGetOrderByHeader); len(orderByVal) > 0 {
		// Check is a field in workflow_dag_result
		isColumn := false
		for _, column := range models.AllDAGResultCols() {
			if models.DAGResultTable + "." + column == orderByVal {
				isColumn = true
				break
			}
		}
		if !isColumn {
			// Check is a field in workflow_dag
			for _, column := range models.AllDAGCols() {
				if models.DagTable + "." + column == orderByVal {
					isColumn = true
					break
				}
			}
			if !isColumn {
				return nil, http.StatusBadRequest, errors.Wrap(err, "Invalid order_by value.")
			}
		}
		orderBy = orderByVal
	}

	return &dagResultsGetArgs{
		AqContext:  aqContext,
		workflowID: workflowID,
		orderBy: orderBy,
		limit: limit,
	}, http.StatusOK, nil
}

func (h *DAGResultsGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*dagResultsGetArgs)

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

	dbDAGResults, err := h.DAGResultRepo.GetByWorkflow(ctx, args.workflowID, args.orderBy, args.limit, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error reading dag results.")
	}

	return slices.Map(dbDAGResults, func(dbResult models.DAGResult) response.DAGResult {
		return *response.NewDAGResultFromDBObject(&dbResult)
	}), http.StatusOK, nil
}
