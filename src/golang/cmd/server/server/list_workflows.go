package server

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// Route: /workflows
// Method: GET
// Params: None
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		serialized `listWorkflowsResponse`, a list of workflow information in the user's org

type workflowResponse struct {
	Id              uuid.UUID              `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	CreatedAt       int64                  `json:"created_at"`
	LastRunAt       int64                  `json:"last_run_at"`
	Status          shared.ExecutionStatus `json:"status"`
	WatcherAuth0Ids []string               `json:"watcher_auth0_id"`
}

type ListWorkflowsHandler struct {
	GetHandler

	Database       database.Database
	WorkflowReader workflow.Reader
}

func (*ListWorkflowsHandler) Name() string {
	return "ListWorkflows"
}

func (*ListWorkflowsHandler) Prepare(r *http.Request) (interface{}, int, error) {
	common, statusCode, err := ParseCommonArgs(r)
	if err != nil {
		return nil, statusCode, err
	}

	return common, http.StatusOK, nil
}

func (h *ListWorkflowsHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*CommonArgs)

	dbWorkflows, err := h.WorkflowReader.GetWorkflowsWithLatestRunResult(ctx, args.OrganizationId, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to list workflows.")
	}

	workflowIds := make([]uuid.UUID, 0, len(dbWorkflows))
	for _, dbWorkflow := range dbWorkflows {
		workflowIds = append(workflowIds, dbWorkflow.Id)
	}

	workflows := make([]workflowResponse, 0, len(dbWorkflows))
	if len(workflowIds) > 0 {
		watchersInfo, err := h.WorkflowReader.GetWatchersInBatch(ctx, workflowIds, h.Database)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to Get Watchers in Batch.")
		}

		watchersMap := make(map[uuid.UUID][]string)
		for _, row := range watchersInfo {
			if _, ok := watchersMap[row.WorkflowId]; !ok {
				watchersMap[row.WorkflowId] = make([]string, 0, len(watchersInfo))
			}
			watchersMap[row.WorkflowId] = append(watchersMap[row.WorkflowId], row.Auth0Id)
		}

		for _, dbWorkflow := range dbWorkflows {
			workflowResponse := workflowResponse{
				Id:              dbWorkflow.Id,
				Name:            dbWorkflow.Name,
				Description:     dbWorkflow.Description,
				CreatedAt:       dbWorkflow.CreatedAt.Unix(),
				LastRunAt:       dbWorkflow.LastRunAt.Unix(),
				Status:          dbWorkflow.Status,
				WatcherAuth0Ids: watchersMap[dbWorkflow.Id],
			}
			workflows = append(workflows, workflowResponse)
		}
	}

	return workflows, http.StatusOK, nil
}
