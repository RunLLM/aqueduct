package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/config"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/logging"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	db_operator "github.com/aqueducthq/aqueduct/lib/models/shared/operator"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
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
	Id              uuid.UUID                 `json:"id"`
	Name            string                    `json:"name"`
	Description     string                    `json:"description"`
	CreatedAt       int64                     `json:"created_at"`
	LastRunAt       int64                     `json:"last_run_at"`
	Status          shared.ExecutionStatus    `json:"status"`
	Engine          shared.EngineType         `json:"engine"`
	OperatorEngines []shared.EngineType       `json:"operator_engines"`
	Checks          []operator.ResultResponse `json:"checks"`
	Metrics         []artifact.ResultResponse `json:"metrics"`
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

	storageConfig := config.Storage()
	vaultObject, err := vault.NewVault(&storageConfig, config.EncryptionKey())
	if err != nil {
		return nil,
			http.StatusInternalServerError,
			errors.Wrap(err, "Unable to initialize vault to sync self-orchestrated workflows.")
	}

	// Asynchronously sync self-orchestrated workflow runs
	go func() {
		if err := engine.SyncSelfOrchestratedWorkflows(
			context.Background(),
			args.OrgID,
			h.ArtifactRepo,
			h.ArtifactResultRepo,
			h.DAGRepo,
			h.DAGEdgeRepo,
			h.DAGResultRepo,
			h.OperatorRepo,
			h.OperatorResultRepo,
			h.WorkflowRepo,
			vaultObject,
			h.Database,
		); err != nil {
			logging.LogAsyncEvent(ctx, logging.ServerComponent, "Sync Workflows", err)
		}
	}()

	latestStatuses, err := h.WorkflowRepo.GetLatestStatusesByOrg(ctx, args.OrgID, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to list workflows.")
	}

	dagIDs := make([]uuid.UUID, 0, len(latestStatuses))
	dagResultIDs := make([]uuid.UUID, 0, len(latestStatuses))
	for _, status := range latestStatuses {
		dagIDs = append(dagIDs, status.DagID)
		dagResultIDs = append(dagResultIDs, status.ResultID)
	}

	engineTypesByDagID, err := h.OperatorRepo.GetEngineTypesMapByDagIDs(ctx, dagIDs, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to get engine types.")
	}

	checkResults, err := h.OperatorResultRepo.GetWithOperatorByDAGResultBatch(
		ctx,
		dagResultIDs,
		[]db_operator.Type{db_operator.CheckType},
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
				Id:              latestStatus.ID,
				Name:            latestStatus.Name,
				Description:     latestStatus.Description,
				CreatedAt:       latestStatus.CreatedAt.Unix(),
				Engine:          latestStatus.Engine,
				OperatorEngines: engineTypesByDagID[latestStatus.DagID],
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
					} else if !errors.Is(err, storage.ErrObjectDoesNotExist()) {
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
				response.Status = shared.RegisteredExecutionStatus
			}

			workflowResponses = append(workflowResponses, response)
		}
	}

	return workflowResponses, http.StatusOK, nil
}
