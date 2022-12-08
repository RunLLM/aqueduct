package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/airflow"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	workflow_utils "github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Route: /workflow/{workflowId}
// Method: GET
// Params:
//	`workflowId`: ID for `workflow` object
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		serialized `getWorkflowResponse`,
//		all metadata and results information for the given `workflowId`

type getWorkflowArgs struct {
	*aq_context.AqContext
	workflowID uuid.UUID
}

type getWorkflowResponse struct {
	// a map of workflow dags keyed by their IDs
	DAGs map[uuid.UUID]*models.DAG `json:"workflow_dags"`
	// a list of dag results. Each result's `workflow_dag_id` field correspond to the
	WorkflowDagResults []workflowDagResult `json:"workflow_dag_results"`
}

type workflowDagResult struct {
	Id            uuid.UUID              `json:"id"`
	CreatedAt     int64                  `json:"created_at"`
	Status        shared.ExecutionStatus `json:"status"`
	WorkflowDagId uuid.UUID              `json:"workflow_dag_id"`
}

type GetWorkflowHandler struct {
	GetHandler

	Database database.Database
	Vault    vault.Vault

	ArtifactReader        artifact.Reader
	OperatorReader        operator.Reader
	WorkflowReader        workflow.Reader
	WorkflowDagReader     workflow_dag.Reader
	WorkflowDagEdgeReader workflow_dag_edge.Reader

	WorkflowDagWriter    workflow_dag.Writer
	OperatorResultWriter operator_result.Writer
	ArtifactResultWriter artifact_result.Writer

	DAGRepo       repos.DAG
	DAGResultRepo repos.DAGResult
}

func (*GetWorkflowHandler) Name() string {
	return "GetWorkflow"
}

func (h *GetWorkflowHandler) Prepare(r *http.Request) (interface{}, int, error) {
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
		aqContext.OrgID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during workflow ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this workflow.")
	}

	return &getWorkflowArgs{
		AqContext:  aqContext,
		workflowID: workflowId,
	}, http.StatusOK, nil
}

func (h *GetWorkflowHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*getWorkflowArgs)

	emptyResp := getWorkflowResponse{}

	latestDAG, err := workflow_utils.ReadLatestDAGFromDatabase(
		ctx,
		args.workflowID,
		h.WorkflowReader,
		h.DAGRepo,
		h.OperatorReader,
		h.ArtifactReader,
		h.WorkflowDagEdgeReader,
		h.Database,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow.")
	}

	if latestDAG.EngineConfig.Type == shared.AirflowEngineType {
		// Airflow workflows need to be synced
		if err := airflow.SyncDAGs(
			ctx,
			[]uuid.UUID{latestDAG.ID},
			h.WorkflowReader,
			h.DAGRepo,
			h.OperatorReader,
			h.ArtifactReader,
			h.WorkflowDagEdgeReader,
			h.DAGResultRepo,
			h.OperatorResultWriter,
			h.ArtifactResultWriter,
			h.Vault,
			h.Database,
		); err != nil {
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow.")
		}
	}

	dbDAGs, err := h.DAGRepo.GetByWorkflow(
		ctx,
		args.workflowID,
		h.Database,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow.")
	}

	dags := make(map[uuid.UUID]*models.DAG, len(dbDAGs))
	for _, dbDAG := range dbDAGs {
		constructedDAG, err := workflow_utils.ReadDAGFromDatabase(
			ctx,
			dbDAG.ID,
			h.WorkflowReader,
			h.DAGRepo,
			h.OperatorReader,
			h.ArtifactReader,
			h.WorkflowDagEdgeReader,
			h.Database,
		)
		if err != nil {
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow.")
		}

		if dbDAG.EngineConfig.Type == shared.AirflowEngineType {
			// TODO: ENG-1714
			// This is a hack for the UI where the `matches_airflow` field
			// for Airflow workflows is set to the value of the latest DAG
			constructedDAG.EngineConfig.AirflowConfig.MatchesAirflow = latestDAG.EngineConfig.AirflowConfig.MatchesAirflow
		}

		dags[dbDAG.ID] = constructedDAG
	}

	dagResults, err := h.DAGResultRepo.GetByWorkflow(ctx, args.workflowID, h.Database)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow.")
	}

	workflowDagResults := make([]workflowDagResult, 0, len(dagResults))
	for _, dagResult := range dagResults {
		workflowDagResults = append(workflowDagResults, workflowDagResult{
			Id:            dagResult.ID,
			CreatedAt:     dagResult.CreatedAt.Unix(),
			Status:        shared.ExecutionStatus(dagResult.Status),
			WorkflowDagId: dagResult.DagID,
		})
	}

	return getWorkflowResponse{
		DAGs:               dags,
		WorkflowDagResults: workflowDagResults,
	}, http.StatusOK, nil
}
