package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/user"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_result"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	workflow_utils "github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

// Route: /workflow/{workflowId}/tables
// Method: GET
// Params:
//	`workflowId`: ID for `workflow` object
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		all tables for the given `workflowId`

type getWorkflowTablesArgs struct {
	*aq_context.AqContext
	workflowId uuid.UUID
}

type getWorkflowTablesResponse struct {
	Tables []Load `json:"tables"`
}

type GetWorkflowTablesHandler struct {
	GetHandler

	Database                database.Database
	// ArtifactReader          artifact.Reader
	// OperatorReader          operator.Reader
	// UserReader              user.Reader
	// WorkflowReader          workflow.Reader
	// WorkflowDagReader       workflow_dag.Reader
	// WorkflowDagEdgeReader   workflow_dag_edge.Reader
	// WorkflowDagResultReader workflow_dag_result.Reader
}

func (*GetWorkflowTablesHandler) Name() string {
	return "GetWorkflowTables"
}

func (h *GetWorkflowTablesHandler) Prepare(r *http.Request) (interface{}, int, error) {
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

	return &getWorkflowTablesArgs{
		AqContext:  aqContext,
		workflowId: workflowId,
	}, http.StatusOK, nil
}

func (h *GetWorkflowTablesHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*getWorkflowArgs)

	emptyResp := getWorkflowTablesResponse{}

	// Get all workflow DAGs for the workflow.
	dbWorkflowDags, err := h.WorkflowDagReader.GetWorkflowDagsByWorkflowId(
		ctx,
		args.workflowId,
		h.Database,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow.")
	}

	loadCount := 0
	for _, dbWorkflowDag := range dbWorkflowDags {
		// Iterate through all operators and count the number of load operators.
		for _, workflowOperator := range dbWorkflowDags.Operators {
			if workflowOperator.Spec.IsLoad() {
				loadCount += 1
			}
		}
	}

	tables := make([]Load, 0, len(loadCount));
	for _, dbWorkflowDag := range dbWorkflowDags {
		// Iterate through all operators and get only load operators.
		for _, workflowOperator := range dbWorkflowDags.Operators {
			if workflowOperator.Spec.IsLoad() {
				loadOperator := workflowOperator.Spec.Load()
				// Add to tables list
				// TODO: Table list is dynamically expanded. How do I create this list? Currently counting # of load ops then adding to the table. Seems like a bit of a waste because two identical for loops.
				table := workflowOperator.Spec.Load()
				tables = append(tables, table)
			}
		}
	}
	// TODO: Test by querying
	// create workflow that loads table into db
	// curl --header "api-key: 1CZR2J96NKXL4UWEO57PID03GAVTHB8Q" http://localhost:8080/api/workflow/{workflow_id}/tables
	return getWorkflowTablesResponse{
		Tables: tables,
	}, http.StatusOK, nil
}
