package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

// Route: /workflow/{workflowId}/objects
// Method: GET
// Params:
//	`workflowId`: ID for `workflow` object
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		all written objects for the given `workflowId`

type GetWorkflowObjectsArgs struct {
	*aq_context.AqContext
	workflowId uuid.UUID
}

type GetWorkflowObjectsResponse struct {
	LoadDetails []operator.GetDistinctLoadOperatorsByWorkflowIdResponse `json:"object_details"`
}

type GetWorkflowObjectsHandler struct {
	GetHandler

	Database       database.Database
	OperatorReader operator.Reader
	WorkflowReader workflow.Reader
}

func (*GetWorkflowObjectsHandler) Name() string {
	return "GetWorkflowObjects"
}

func (h *GetWorkflowObjectsHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	workflowIdStr := chi.URLParam(r, routes.WorkflowIdUrlParam)
	workflowId, err := uuid.Parse(workflowIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed workflow ID.")
	}

	ok, err := h.WorkflowReader.ValidateWorkflowOwnership(
		r.Context(),
		workflowId,
		aqContext.OrganizationId,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during workflow ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this workflow.")
	}

	return &GetWorkflowObjectsArgs{
		AqContext:  aqContext,
		workflowId: workflowId,
	}, http.StatusOK, nil
}

func (h *GetWorkflowObjectsHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*GetWorkflowObjectsArgs)

	emptyResp := GetWorkflowObjectsResponse{}

	// Get all specs  for the workflow.
	operatorList, err := h.OperatorReader.GetDistinctLoadOperatorsByWorkflowId(ctx, args.workflowId, h.Database)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow.")
	}

	return GetWorkflowObjectsResponse{
		LoadDetails: operatorList,
	}, http.StatusOK, nil
}
