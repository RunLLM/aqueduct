package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/workflow/dag"
	workflow_utils "github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Route: /workflow/{workflowId}/dag/{workflowDagId}
// Method: GET
// Params:
//	`workflowId`: ID for `workflow` object
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		serialized `dag.Response`

type getWorkflowDagArgs struct {
	*aq_context.AqContext
	workflowID uuid.UUID
	dagID      uuid.UUID
}

type GetWorkflowDAGHandler struct {
	GetHandler

	Database database.Database

	ArtifactRepo repos.Artifact
	DAGRepo      repos.DAG
	DAGEdgeRepo  repos.DAGEdge
	OperatorRepo repos.Operator
	WorkflowRepo repos.Workflow
}

func (*GetWorkflowDAGHandler) Name() string {
	return "GetWorkflowDAG"
}

func (h *GetWorkflowDAGHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	workflowIDStr := chi.URLParam(r, routes.WorkflowIdUrlParam)
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed workflow ID.")
	}

	dagIDStr := chi.URLParam(r, routes.WorkflowDagIDUrlParam)
	dagID, err := uuid.Parse(dagIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed workflow DAG ID.")
	}

	ok, err := h.WorkflowRepo.ValidateOrg(
		r.Context(),
		workflowID,
		aqContext.OrgID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during workflow ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this workflow.")
	}

	return &getWorkflowDagArgs{
		AqContext:  aqContext,
		dagID:      dagID,
		workflowID: workflowID,
	}, http.StatusOK, nil
}

func (h *GetWorkflowDAGHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*getWorkflowDagArgs)

	emptyResp := dag.ResultResponse{}

	// Read dag structure
	constructedDAG, err := workflow_utils.ReadDAGFromDatabase(
		ctx,
		args.dagID,
		h.WorkflowRepo,
		h.DAGRepo,
		h.OperatorRepo,
		h.ArtifactRepo,
		h.DAGEdgeRepo,
		h.Database,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow.")
	}

	return dag.NewResponseFromDbObjects(constructedDAG), http.StatusOK, nil
}
