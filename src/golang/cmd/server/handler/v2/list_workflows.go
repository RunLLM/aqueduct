package v2

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/response"
	"github.com/dropbox/godropbox/errors"
)

/*
This file should map directly to src/ui/common/src/handlers/ListWorkflows.tsx

Route: /v2/workflows
Method: GET
Request:
	Headers:
		`api-key`:
			User's API Key
Response:
	Body:
		List of `response.Workflow` objects
*/

type ListWorkflowsHandler struct {
	handler.GetHandler

	Database database.Database

	WorkflowRepo repos.Workflow
}

func (*ListWorkflowsHandler) Name() string {
	return "ListWorkflows"
}

func (h *ListWorkflowsHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	return &workflowGetArgs{
		AqContext: aqContext,
	}, http.StatusOK, nil
}

func (h *ListWorkflowsHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	dbWorkflows, err := h.WorkflowRepo.List(
		ctx,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during the retrieval of workflows.")
	}

	workflows := make([]*response.Workflow, len(dbWorkflows))

	for idx, dbWorkflow := range dbWorkflows {
		workflows[idx] = response.NewWorkflowFromDBObject(&dbWorkflow)
	}

	return workflows, http.StatusOK, nil
}
