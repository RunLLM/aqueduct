package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	db_artifact "github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Route: /artifact/{artifactId}/results
// Method: GET
// Params: None
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		serialized `listArtifactResultsResponse`

type listArtifactResultsResponse struct {
	Results []artifact.RawResultResponse `json:"results"`
}

type listArtifactResultsArgs struct {
	*aq_context.AqContext
	ArtifactId uuid.UUID
}

type ListArtifactResultsHandler struct {
	GetHandler

	Database             database.Database
	ArtifactReader       db_artifact.Reader
	ArtifactResultReader artifact_result.Reader
	WorkflowDagReader    workflow_dag.Reader
}

func (*ListArtifactResultsHandler) Name() string {
	return "ListArtifactResults"
}

func (*ListArtifactResultsHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	artfIdStr := chi.URLParam(r, routes.ArtifactIdUrlParam)
	artfId, err := uuid.Parse(artfIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed artifact ID.")
	}

	return &listArtifactResultsArgs{
		AqContext:  aqContext,
		ArtifactId: artfId,
	}, http.StatusOK, nil
}

func (h *ListArtifactResultsHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*listArtifactResultsArgs)
	artfId := args.ArtifactId

	emptyResponse := listArtifactResultsResponse{}

	results, err := h.ArtifactResultReader.GetArtifactResultsByArtifactId(ctx, artfId, h.Database)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve artifact results.")
	}

	artf, err := h.ArtifactReader.GetArtifact(ctx, artfId, h.Database)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve artifact.")
	}

	if len(results) == 0 {
		return emptyResponse, http.StatusOK, nil
	}

	resultIds := make([]uuid.UUID, 0, len(results))
	for _, result := range results {
		resultIds = append(resultIds, result.Id)
	}

	dbDagsByResultIds, err := h.WorkflowDagReader.GetWorkflowDagsMapByArtifactResultIds(ctx, resultIds, h.Database)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve workflow dags.")
	}

	// maps from db dag Ids
	dbDagsMapByDagIds := make(map[uuid.UUID]workflow_dag.DBWorkflowDag, len(dbDagsByResultIds))
	artfResultsMapByDagIds := make(map[uuid.UUID][]artifact_result.ArtifactResult, len(dbDagsByResultIds))
	for _, artfResult := range results {
		if dbDag, ok := dbDagsByResultIds[artfResult.Id]; ok {
			if _, okDagsMap := dbDagsMapByDagIds[dbDag.Id]; !okDagsMap {
				dbDagsMapByDagIds[dbDag.Id] = dbDag
			}

			artfResultsMapByDagIds[dbDag.Id] = append(artfResultsMapByDagIds[dbDag.Id], artfResult)
		} else {
			return emptyResponse, http.StatusInternalServerError, errors.Newf("Error retrieving dag associated with artifact result %s", artfResult.Id)
		}
	}

	responses := make([]artifact.RawResultResponse, 0, len(results))
	for dbDagId, artfResults := range artfResultsMapByDagIds {
		if dag, ok := dbDagsMapByDagIds[dbDagId]; ok {
			storageObj := storage.NewStorage(&dag.StorageConfig)
			if err != nil {
				return emptyResponse, http.StatusInternalServerError, errors.New("Error retrieving artifact contents.")
			}

			for _, artfResult := range artfResults {
				var contentPtr *string = nil
				if artf.Type.IsCompact() && !artfResult.ExecState.IsNull && artfResult.ExecState.ExecutionState.Terminated() {
					contentBytes, err := storageObj.Get(ctx, artfResult.ContentPath)
					if err != nil {
						return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, fmt.Sprintf("Error retrieving artifact content for result %s", artfResult.Id))
					}

					contentStr := string(contentBytes)
					contentPtr = &contentStr
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
