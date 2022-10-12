package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/queries"
	"github.com/aqueducthq/aqueduct/lib/airflow"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/user"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_result"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/logging"
	"github.com/aqueducthq/aqueduct/lib/vault"
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
	Engine          string                 `json:"engine"`
	WatcherAuth0Ids []string               `json:"watcher_auth0_id"`
}

type ListWorkflowsHandler struct {
	GetHandler

	Database database.Database
	Vault    vault.Vault

	UserReader              user.Reader
	ArtifactReader          artifact.Reader
	OperatorReader          operator.Reader
	WorkflowReader          workflow.Reader
	WorkflowDagReader       workflow_dag.Reader
	WorkflowDagEdgeReader   workflow_dag_edge.Reader
	WorkflowDagResultReader workflow_dag_result.Reader
	CustomReader            queries.Reader

	ArtifactWriter          artifact.Writer
	OperatorWriter          operator.Writer
	WorkflowWriter          workflow.Writer
	WorkflowDagWriter       workflow_dag.Writer
	WorkflowDagEdgeWriter   workflow_dag_edge.Writer
	WorkflowDagResultWriter workflow_dag_result.Writer
	OperatorResultWriter    operator_result.Writer
	ArtifactResultWriter    artifact_result.Writer
	NotificationWriter      notification.Writer
}

func (*ListWorkflowsHandler) Name() string {
	return "ListWorkflows"
}

func (*ListWorkflowsHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	return aqContext, http.StatusOK, nil
}

func (h *ListWorkflowsHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*aq_context.AqContext)

	// Asynchronously sync self-orchestrated workflow runs
	go func() {
		if err := syncSelfOrchestratedWorkflows(context.Background(), h, args.OrganizationId); err != nil {
			logging.LogAsyncEvent(ctx, logging.ServerComponent, "Sync Workflows", err)
		}
	}()

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
			response := workflowResponse{
				Id:              dbWorkflow.Id,
				Name:            dbWorkflow.Name,
				Description:     dbWorkflow.Description,
				CreatedAt:       dbWorkflow.CreatedAt.Unix(),
				Engine:          dbWorkflow.Engine,
				WatcherAuth0Ids: watchersMap[dbWorkflow.Id],
			}
			if !dbWorkflow.LastRunAt.IsNull {
				response.LastRunAt = dbWorkflow.LastRunAt.Time.Unix()
			}

			if !dbWorkflow.Status.IsNull {
				response.Status = dbWorkflow.Status.ExecutionStatus
			} else {
				// There are no workflow runs yet for this workflow, so we simply return
				// that the workflow has been registered
				response.Status = shared.RegisteredExecutionStatus
			}
			workflows = append(workflows, response)
		}
	}
	return workflows, http.StatusOK, nil
}

// syncSelfOrchestratedWorkflows syncs any workflow DAG results for any workflows running on a
// self-orchestrated engine for the user's organization.
func syncSelfOrchestratedWorkflows(ctx context.Context, h *ListWorkflowsHandler, organizationID string) error {
	// Sync workflows running on self-orchestrated engines
	airflowWorkflowDagIds, err := h.CustomReader.GetLatestWorkflowDagIdsByOrganizationIdAndEngine(
		ctx,
		organizationID,
		shared.AirflowEngineType,
		h.Database,
	)
	if err != nil {
		return err
	}

	airflowWorkflowDagUUIDs := make([]uuid.UUID, 0, len(airflowWorkflowDagIds))
	for _, workflowDagId := range airflowWorkflowDagIds {
		airflowWorkflowDagUUIDs = append(airflowWorkflowDagUUIDs, workflowDagId.Id)
	}

	if err := airflow.SyncWorkflowDags(
		ctx,
		airflowWorkflowDagUUIDs,
		h.WorkflowReader,
		h.WorkflowDagReader,
		h.OperatorReader,
		h.ArtifactReader,
		h.WorkflowDagEdgeReader,
		h.WorkflowDagResultReader,
		h.WorkflowDagWriter,
		h.WorkflowDagResultWriter,
		h.OperatorResultWriter,
		h.ArtifactResultWriter,
		h.Vault,
		h.Database,
	); err != nil {
		return err
	}

	return nil
}
