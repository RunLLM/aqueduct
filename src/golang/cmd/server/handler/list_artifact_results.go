package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Route: /workflow/{workflowId}/artifact/{artifactId}/results
// Method: GET
// Params: None
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		serialized `listArtifactResultsResponse`

type listArtifactResultsResponse struct {
	// Results are not ordered.
	Results []artifact.RawResultResponse `json:"results"`
}

type listArtifactResultsArgs struct {
	*aq_context.AqContext
	WorkflowId uuid.UUID
	ArtifactId uuid.UUID
}

type ListArtifactResultsHandlerDeprecated struct {
	GetHandler

	Database database.Database

	ArtifactRepo       repos.Artifact
	ArtifactResultRepo repos.ArtifactResult
	DAGRepo            repos.DAG
}

func (*ListArtifactResultsHandlerDeprecated) Name() string {
	return "ListArtifactResults"
}

func (*ListArtifactResultsHandlerDeprecated) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	artfIdStr := chi.URLParam(r, routes.ArtifactIdUrlParam)
	artfId, err := uuid.Parse(artfIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed artifact ID.")
	}

	wfIdStr := chi.URLParam(r, routes.WorkflowIdUrlParam)
	wfId, err := uuid.Parse(wfIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed workflow Id.")
	}

	return &listArtifactResultsArgs{
		AqContext:  aqContext,
		ArtifactId: artfId,
		WorkflowId: wfId,
	}, http.StatusOK, nil
}

func (h *ListArtifactResultsHandlerDeprecated) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*listArtifactResultsArgs)
	artfID := args.ArtifactId
	wfID := args.WorkflowId

	emptyResponse := listArtifactResultsResponse{}

	artf, err := h.ArtifactRepo.Get(ctx, artfID, h.Database)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve artifact.")
	}

	results, err := h.ArtifactResultRepo.GetByArtifactNameAndWorkflow(ctx, artf.Name, wfID, h.Database)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve artifact results.")
	}

	if len(results) == 0 {
		return emptyResponse, http.StatusOK, nil
	}

	resultIds := make([]uuid.UUID, 0, len(results))
	for _, result := range results {
		resultIds = append(resultIds, result.ID)
	}

	artfResultToDAG, err := h.DAGRepo.GetByArtifactResultBatch(ctx, resultIds, h.Database)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve workflow dags.")
	}

	// maps from db dag Ids
	dbDagByDagId := make(map[uuid.UUID]models.DAG, len(artfResultToDAG))
	artfResultByDagId := make(map[uuid.UUID][]models.ArtifactResult, len(artfResultToDAG))
	for _, artfResult := range results {
		if dbDag, ok := artfResultToDAG[artfResult.ID]; ok {
			if _, okDagsMap := dbDagByDagId[dbDag.ID]; !okDagsMap {
				dbDagByDagId[dbDag.ID] = dbDag
			}

			artfResultByDagId[dbDag.ID] = append(artfResultByDagId[dbDag.ID], artfResult)
		} else {
			return emptyResponse, http.StatusInternalServerError, errors.Newf("Error retrieving dag associated with artifact result %s", artfResult.ID)
		}
	}

	responses := make([]artifact.RawResultResponse, 0, len(results))
	for dbDagId, artfResults := range artfResultByDagId {
		if dag, ok := dbDagByDagId[dbDagId]; ok {
			storageObj := storage.NewStorage(&dag.StorageConfig)
			if err != nil {
				return emptyResponse, http.StatusInternalServerError, errors.New("Error retrieving artifact contents.")
			}

			for _, artfResult := range artfResults {
				var contentPtr *string = nil
				if artf.Type.IsCompact() &&
					!artfResult.ExecState.IsNull &&
					(artfResult.ExecState.ExecutionState.Status == shared.FailedExecutionStatus ||
						artfResult.ExecState.ExecutionState.Status == shared.SucceededExecutionStatus) {
					exists := storageObj.Exists(ctx, artfResult.ContentPath)
					if exists {
						contentBytes, err := storageObj.Get(ctx, artfResult.ContentPath)
						if err != nil {
							return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, fmt.Sprintf("Error retrieving artifact content for result %s", artfResult.ID))
						}

						contentStr := string(contentBytes)
						contentPtr = &contentStr
					}
				}

				responses = append(responses, *artifact.NewRawResultResponseFromDbObject(
					&artfResult, contentPtr,
				))
			}
		} else {
			return emptyResponse, http.StatusInternalServerError, errors.Newf("Error retrieving dag %s", dbDagId)
		}
	}

	return &listArtifactResultsResponse{Results: responses}, http.StatusOK, nil
}
