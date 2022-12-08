package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/aqueducthq/aqueduct/lib/workflow/dag"
	workflow_utils "github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Route: /workflow/{workflowDagId}/result/{workflowDagResultId}
// Method: GET
// Params:
//	`workflowId`: ID for `workflow` object
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		serialized `dag.ResultResponse`

type getWorkflowDagResultArgs struct {
	*aq_context.AqContext
	workflowID  uuid.UUID
	dagResultID uuid.UUID
}

type GetWorkflowDagResultHandler struct {
	GetHandler

	Database             database.Database
	ArtifactResultReader artifact_result.Reader
	OperatorReader       operator.Reader
	OperatorResultReader operator_result.Reader

	ArtifactRepo  repos.Artifact
	DAGRepo       repos.DAG
	DAGEdgeRepo   repos.DAGEdge
	DAGResultRepo repos.DAGResult
	WorkflowRepo  repos.Workflow
}

func (*GetWorkflowDagResultHandler) Name() string {
	return "GetWorkflowDagResult"
}

func (h *GetWorkflowDagResultHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	workflowIDStr := chi.URLParam(r, routes.WorkflowIdUrlParam)
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed workflow ID.")
	}

	dagResultIDStr := chi.URLParam(r, routes.WorkflowDagResultIdUrlParam)
	dagResultID, err := uuid.Parse(dagResultIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed workflow dag result ID.")
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

	return &getWorkflowDagResultArgs{
		AqContext:   aqContext,
		workflowID:  workflowID,
		dagResultID: dagResultID,
	}, http.StatusOK, nil
}

func (h *GetWorkflowDagResultHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*getWorkflowDagResultArgs)

	emptyResp := dag.ResultResponse{}

	dbDAG, err := h.DAGRepo.GetByDAGResult(
		ctx,
		args.dagResultID,
		h.Database,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow dag.")
	}

	// Read dag structure
	constructedDAG, err := workflow_utils.ReadDAGFromDatabase(
		ctx,
		dbDAG.ID,
		h.WorkflowRepo,
		h.DAGRepo,
		h.OperatorReader,
		h.ArtifactRepo,
		h.DAGEdgeRepo,
		h.Database,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow.")
	}

	dagResult, err := h.DAGResultRepo.Get(ctx, args.dagResultID, h.Database)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow.")
	}

	operatorResults, err := h.OperatorResultReader.GetOperatorResultsByWorkflowDagResultIds(
		ctx,
		[]uuid.UUID{args.dagResultID},
		h.Database,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving operator results.")
	}

	artifactResults, err := h.ArtifactResultReader.GetArtifactResultsByWorkflowDagResultIds(
		ctx,
		[]uuid.UUID{args.dagResultID},
		h.Database,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving artifact results.")
	}

	contents, err := getArtifactContents(ctx, constructedDAG, artifactResults)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving artifact contents.")
	}

	return dag.NewResultResponseFromDbObjects(
		constructedDAG,
		dagResult,
		operatorResults,
		artifactResults,
		contents,
	), http.StatusOK, nil
}

// TODO: We should replace this once we migrate to services based on `artifact` objects
// Fetches the artifact contents for all compact artifact data types. Returns a map from
// content path to content. If an artifact's data was never written, it's entry will be
// excluded from the map.
func getArtifactContents(
	ctx context.Context,
	dag *models.DAG,
	dbArtifactResults []artifact_result.ArtifactResult,
) (map[string]string, error) {
	contents := make(map[string]string, len(dbArtifactResults))
	storageObj := storage.NewStorage(&dag.StorageConfig)
	for _, artfResult := range dbArtifactResults {
		if artf, ok := dag.Artifacts[artfResult.ArtifactId]; ok {
			// These artifacts has small content size and we can safely include them all in response.
			if artf.Type.IsCompact() {
				path := artfResult.ContentPath
				// Read data from storage and deserialize payload to `container`.
				contentBytes, err := storageObj.Get(ctx, path)
				if err == storage.ErrObjectDoesNotExist {
					// If the data does not exist, skip the fetch.
					continue
				}
				if err != nil {
					return nil, errors.Wrap(err, "Unable to get artifact content from storage")
				}
				contents[path] = string(contentBytes)
			}
		}
	}

	return contents, nil
}
