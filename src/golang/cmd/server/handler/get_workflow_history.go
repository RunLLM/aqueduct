package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Route: /workflow/{workflowId}/history
// Method: GET
// Params:
//  `workflowId`: the UUID of the workflow whose history we're retrieving
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		serialized `workflowHistoryResponse`, a list of all the versions of this workflow, their timestamps, and their statuses

type getWorkflowHistoryResponse struct {
	Id   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	// The versions here will be returned in order of creation. The oldest version will be first and the most recent version will be
	// last.
	Versions []workflowVersionResponse `json:"versions"`
}

type workflowVersionResponse struct {
	Id        uuid.UUID                 `json:"versionId"`
	CreatedAt int64                     `json:"created_at"`
	ExecState shared.NullExecutionState `json:"exec_state"`
}

type getWorkflowHistoryArgs struct {
	*aq_context.AqContext
	workflowId uuid.UUID
}

type GetWorkflowHistoryHandler struct {
	GetHandler

	Database database.Database

	DAGResultRepo repos.DAGResult
	WorkflowRepo  repos.Workflow
}

func (*GetWorkflowHistoryHandler) Name() string {
	return "GetWorkflowHistory"
}

func (h *GetWorkflowHistoryHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	workflowIdStr := chi.URLParam(r, routes.WorkflowIdUrlParam)
	workflowId, err := uuid.Parse(workflowIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed workflow ID.")
	}

	ok, err := h.WorkflowRepo.ValidateOrg(
		r.Context(),
		workflowId,
		aqContext.OrgID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during workflow ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this workflow.")
	}

	return &getWorkflowHistoryArgs{
		AqContext:  aqContext,
		workflowId: workflowId,
	}, http.StatusOK, nil
}

func (h *GetWorkflowHistoryHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*getWorkflowHistoryArgs)

	workflow, err := h.WorkflowRepo.Get(ctx, args.workflowId, h.Database)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, fmt.Sprintf("Workflow %v does not exist.", args.workflowId))
	}

	// Default values to not have an order and not have a limit: Empty string for order_by, -1 for limit
	// Set true for order_by order (desc/asc) because doesn't matter.
	results, err := h.DAGResultRepo.GetByWorkflow(ctx, args.workflowId, "", -1, true, h.Database)
	if err != nil && err != database.ErrNoRows() { // Don't return an error if there are just no rows.
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error while retrieving workflow runs.")
	}

	versions := make([]workflowVersionResponse, len(results))
	for idx, result := range results {
		version := workflowVersionResponse{
			Id:        result.ID,
			CreatedAt: result.CreatedAt.Unix(),
			ExecState: result.ExecState,
		}

		versions[idx] = version
	}

	return getWorkflowHistoryResponse{
		Id:       workflow.ID,
		Name:     workflow.Name,
		Versions: versions,
	}, http.StatusOK, nil
}
