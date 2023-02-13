package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/airflow"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/logging"
	mdl_shared "github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
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
	Id          uuid.UUID                  `json:"id"`
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	CreatedAt   int64                      `json:"created_at"`
	LastRunAt   int64                      `json:"last_run_at"`
	Status      mdl_shared.ExecutionStatus `json:"status"`
	Engine      string                     `json:"engine"`
	Checks      []operator.ResultResponse  `json:"checks"`
	Metrics     []artifact.ResultResponse  `json:"metrics"`
}

type ListWorkflowsHandler struct {
	GetHandler

	Database database.Database

	ArtifactRepo       repos.Artifact
	ArtifactResultRepo repos.ArtifactResult
	DAGRepo            repos.DAG
	DAGEdgeRepo        repos.DAGEdge
	DAGResultRepo      repos.DAGResult
	OperatorRepo       repos.Operator
	OperatorResultRepo repos.OperatorResult
	WorkflowRepo       repos.Workflow
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

	dagResultIDs := make([]uuid.UUID, 0, len(latestStatuses))
	for _, status := range latestStatuses {
		dagResultIDs = append(dagResultIDs, status.ResultID)
	}

	checkResults, err := h.OperatorResultRepo.GetWithOperatorByDAGResultBatch(
		ctx,
		dagResultIDs,
		[]mdl_shared.OperatorType{mdl_shared.CheckType},
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to get checks.")
	}

	checkResultsByDAGResultID := make(map[uuid.UUID][]views.OperatorWithResult, len(checkResults))
	for _, checkResult := range checkResults {
		checkResultsByDAGResultID[checkResult.DAGResultID] = append(
			checkResultsByDAGResultID[checkResult.DAGResultID],
			checkResult,
		)
	}

	metricResults, err := h.ArtifactResultRepo.GetWithArtifactOfMetricsByDAGResultBatch(
		ctx,
		dagResultIDs,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to get metrics.")
	}

	metricResultsByDAGResultID := make(map[uuid.UUID][]views.ArtifactWithResult, len(metricResults))
	for _, metricResult := range metricResults {
		metricResultsByDAGResultID[metricResult.DAGResultID] = append(
			metricResultsByDAGResultID[metricResult.DAGResultID],
			metricResult,
		)
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

			for _, checkResult := range checkResultsByDAGResultID[latestStatus.ResultID] {
				response.Checks = append(
					response.Checks,
					*operator.NewResultResponseFromDBView(&checkResult),
				)
			}

			for _, metricResult := range metricResultsByDAGResultID[latestStatus.ResultID] {
				var contentPtr *string = nil
				if metricResult.Type.IsCompact() {
					storageObj := storage.NewStorage(&metricResult.StorageConfig)
					path := metricResult.ContentPath
					contentBytes, err := storageObj.Get(ctx, path)
					if err == nil {
						content := string(contentBytes)
						contentPtr = &content
					} else if err != storage.ErrObjectDoesNotExist {
						return nil, http.StatusInternalServerError, errors.Wrap(
							err, "Unable to get metric content from storage",
						)
					}
				}

				response.Metrics = append(
					response.Metrics,
					*artifact.NewResultResponseFromDBView(&metricResult, contentPtr),
				)
			}

			if !latestStatus.LastRunAt.IsNull {
				response.LastRunAt = latestStatus.LastRunAt.Time.Unix()
			}

			if !latestStatus.Status.IsNull {
				response.Status = latestStatus.Status.ExecutionStatus
			} else {
				// There are no workflow runs yet for this workflow, so we simply return
				// that the workflow has been registered
				response.Status = mdl_shared.RegisteredExecutionStatus
			}

			workflowResponses = append(workflowResponses, response)
		}
	}

	return workflowResponses, http.StatusOK, nil
}

// syncSelfOrchestratedWorkflows syncs any workflow DAG results for any workflows running on a
// self-orchestrated engine for the user's organization.
func syncSelfOrchestratedWorkflows(ctx context.Context, h *ListWorkflowsHandler, orgID string) error {
	// Sync workflows running on self-orchestrated engines
	airflowDagIDs, err := h.DAGRepo.GetLatestIDsByOrgAndEngine(
		ctx,
		orgID,
		shared.AirflowEngineType,
		h.Database,
	)
	if err != nil {
		return err
	}

	storageConfig := config.Storage()
	vaultObject, err := vault.NewVault(&storageConfig, config.EncryptionKey())
	if err != nil {
		return errors.Wrap(err, "Unable to initialize vault.")
	}

	if err := airflow.SyncDAGs(
		ctx,
		airflowDagIDs,
		h.WorkflowRepo,
		h.DAGRepo,
		h.OperatorRepo,
		h.ArtifactRepo,
		h.DAGEdgeRepo,
		h.DAGResultRepo,
		h.OperatorResultRepo,
		h.ArtifactResultRepo,
		vaultObject,
		h.Database,
	); err != nil {
		return err
	}

	return nil
}
