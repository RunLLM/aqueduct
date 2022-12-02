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
	"github.com/aqueducthq/aqueduct/lib/collections/user"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_result"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
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
	workflowId uuid.UUID
}

type getWorkflowResponse struct {
	// a map of workflow dags keyed by their IDs
	WorkflowDags map[uuid.UUID]*workflow_dag.DBWorkflowDag `json:"workflow_dags"`
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

	ArtifactReader          artifact.Reader
	OperatorReader          operator.Reader
	UserReader              user.Reader
	WorkflowReader          workflow.Reader
	WorkflowDagReader       workflow_dag.Reader
	WorkflowDagEdgeReader   workflow_dag_edge.Reader
	WorkflowDagResultReader workflow_dag_result.Reader

	WorkflowDagWriter       workflow_dag.Writer
	WorkflowDagResultWriter workflow_dag_result.Writer
	OperatorResultWriter    operator_result.Writer
	ArtifactResultWriter    artifact_result.Writer
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
		aqContext.OrganizationId,
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
		workflowId: workflowId,
	}, http.StatusOK, nil
}

func (h *GetWorkflowHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*getWorkflowArgs)

	emptyResp := getWorkflowResponse{}

	latestWorkflowDag, err := workflow_utils.ReadLatestWorkflowDagFromDatabase(
		ctx,
		args.workflowId,
		h.WorkflowReader,
		h.WorkflowDagReader,
		h.OperatorReader,
		h.ArtifactReader,
		h.WorkflowDagEdgeReader,
		h.Database,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow.")
	}

	if latestWorkflowDag.EngineConfig.Type == shared.AirflowEngineType {
		// Airflow workflows need to be synced
		if err := airflow.SyncWorkflowDags(
			ctx,
			[]uuid.UUID{latestWorkflowDag.Id},
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
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow.")
		}
	}

	dbWorkflowDags, err := h.WorkflowDagReader.GetWorkflowDagsByWorkflowId(
		ctx,
		args.workflowId,
		h.Database,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow.")
	}

	workflowDags := make(map[uuid.UUID]*workflow_dag.DBWorkflowDag, len(dbWorkflowDags))
	for _, dbWorkflowDag := range dbWorkflowDags {
		constructedDag, err := workflow_utils.ReadWorkflowDagFromDatabase(
			ctx,
			dbWorkflowDag.Id,
			h.WorkflowReader,
			h.WorkflowDagReader,
			h.OperatorReader,
			h.ArtifactReader,
			h.WorkflowDagEdgeReader,
			h.Database,
		)
		if err != nil {
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow.")
		}

		if dbWorkflowDag.EngineConfig.Type == shared.AirflowEngineType {
			// TODO: ENG-1714
			// This is a hack for the UI where the `matches_airflow` field
			// for Airflow workflows is set to the value of the latest DAG
			constructedDag.EngineConfig.AirflowConfig.MatchesAirflow = latestWorkflowDag.EngineConfig.AirflowConfig.MatchesAirflow
		}

		workflowDags[dbWorkflowDag.Id] = constructedDag
	}

	dbWorkflowDagResults, err := h.WorkflowDagResultReader.GetWorkflowDagResultsByWorkflowId(
		ctx,
		args.workflowId,
		h.Database,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow.")
	}

	workflowDagResults := make([]workflowDagResult, 0, len(dbWorkflowDagResults))
	for _, dbWorkflowDagResult := range dbWorkflowDagResults {
		workflowDagResults = append(workflowDagResults, workflowDagResult{
			Id:            dbWorkflowDagResult.Id,
			CreatedAt:     dbWorkflowDagResult.CreatedAt.Unix(),
			Status:        dbWorkflowDagResult.Status,
			WorkflowDagId: dbWorkflowDagResult.WorkflowDagId,
		})
	}

	return getWorkflowResponse{
		WorkflowDags:       workflowDags,
		WorkflowDagResults: workflowDagResults,
	}, http.StatusOK, nil
}
