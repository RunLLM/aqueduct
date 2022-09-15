package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Route: /workflow/{artifactId}/results
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
	ArtifactReader       artifact.Reader
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

	artfIdStr := chi.URLParam(r, routes.IntegrationIdUrlParam)
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
	args := interfaceArgs.(*aq_context.AqContext)
	artfId := args.ArtifactId

	emptyResponse := listArtifactResultsResponse{}

	artf, err := h.ArtifactReader.GetArtifact(ctx, artfId)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve artifact.")
	}

	dbDag, err := h.WorkflowDagReader.GetWorkflowDag()

	results, err := h.ArtifactResultReader.GetArtifactResultsByArtifactId(ctx, args.ArtifactId)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve artifact results.")
	}

}
