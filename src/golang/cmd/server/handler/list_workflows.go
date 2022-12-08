package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/queries"
	"github.com/aqueducthq/aqueduct/lib/airflow"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/logging"
	"github.com/aqueducthq/aqueduct/lib/repos"
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
	Id          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	CreatedAt   int64                  `json:"created_at"`
	LastRunAt   int64                  `json:"last_run_at"`
	Status      shared.ExecutionStatus `json:"status"`
	Engine      string                 `json:"engine"`
}

type ListWorkflowsHandler struct {
	GetHandler

	Database database.Database
	Vault    vault.Vault

	ArtifactReader        artifact.Reader
	OperatorReader        operator.Reader
	WorkflowDagEdgeReader workflow_dag_edge.Reader
	CustomReader          queries.Reader

	ArtifactWriter       artifact.Writer
	OperatorWriter       operator.Writer
	OperatorResultWriter operator_result.Writer
	ArtifactResultWriter artifact_result.Writer

	DAGRepo       repos.DAG
	DAGEdgeRepo   repos.DAGEdge
	DAGResultRepo repos.DAGResult
	WorkflowRepo  repos.Workflow
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
		if err := syncSelfOrchestratedWorkflows(context.Background(), h, args.OrgID); err != nil {
			logging.LogAsyncEvent(ctx, logging.ServerComponent, "Sync Workflows", err)
		}
	}()

	latestStatuses, err := h.WorkflowRepo.GetLatestStatusesByOrg(ctx, args.OrgID, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to list workflows.")
	}

	workflowIDs := make([]uuid.UUID, 0, len(latestStatuses))
	for _, latestStatus := range latestStatuses {
		workflowIDs = append(workflowIDs, latestStatus.ID)
	}

	workflowResponses := make([]workflowResponse, 0, len(latestStatuses))
	if len(workflowIDs) > 0 {
		for _, latestStatus := range latestStatuses {
			response := workflowResponse{
				Id:          latestStatus.ID,
				Name:        latestStatus.Name,
				Description: latestStatus.Description,
				CreatedAt:   latestStatus.CreatedAt.Unix(),
				Engine:      latestStatus.Engine,
			}

			if !latestStatus.LastRunAt.IsNull {
				response.LastRunAt = latestStatus.LastRunAt.Time.Unix()
			}

			if !latestStatus.Status.IsNull {
				response.Status = latestStatus.Status.ExecutionStatus
			} else {
				// There are no workflow runs yet for this workflow, so we simply return
				// that the workflow has been registered
				response.Status = shared.RegisteredExecutionStatus
			}

			workflowResponses = append(workflowResponses, response)
		}
	}

	return workflowResponses, http.StatusOK, nil
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

	if err := airflow.SyncDAGs(
		ctx,
		airflowWorkflowDagUUIDs,
		h.WorkflowRepo,
		h.DAGRepo,
		h.OperatorReader,
		h.ArtifactReader,
		h.DAGEdgeRepo,
		h.DAGResultRepo,
		h.OperatorResultWriter,
		h.ArtifactResultWriter,
		h.Vault,
		h.Database,
	); err != nil {
		return err
	}

	return nil
}
