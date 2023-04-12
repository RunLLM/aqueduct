package v2

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/cmd/server/request/parser"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/funcitonal_primitives/functional_map"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/response"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/google/uuid"
)

// This file should map directly to
// src/ui/common/src/handlers/NodeResultsGet.tsx
//
// Route: /v2/workflow/{workflowId}/result/{dagResultID}/nodes/results
// Method: GET
// Params:
//	`workflowId`: ID for `workflow` object
//  `dagResultID`: ID for `workflow_dag_result` object
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		serialized `response.NodeResults`

type nodesResultGetArgs struct {
	*aq_context.AqContext
	workflowID  uuid.UUID
	dagResultID uuid.UUID
}

type NodesResultsGetHandler struct {
	handler.GetHandler

	Database database.Database

	WorkflowRepo       repos.Workflow
	DAGRepo            repos.DAG
	ArtifactRepo       repos.Artifact
	OperatorResultRepo repos.OperatorResult
	ArtifactResultRepo repos.ArtifactResult
}

func (*NodesResultsGetHandler) Name() string {
	return "DAGResultGet"
}

func (h *NodesResultsGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	workflowID, err := (parser.WorkflowIDParser{}).Parse(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	dagResultID, err := (parser.DAGResultIDParser{}).Parse(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return &nodesResultGetArgs{
		AqContext:   aqContext,
		workflowID:  workflowID,
		dagResultID: dagResultID,
	}, http.StatusOK, nil
}

func (h *NodesResultsGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*nodesResultGetArgs)

	ok, err := h.WorkflowRepo.ValidateOrg(
		ctx,
		args.workflowID,
		args.OrgID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during workflow ownership validation.")
	}

	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this workflow.")
	}

	dbDAG, err := h.DAGRepo.GetByDAGResult(ctx, args.dagResultID, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error reading DAG.")
	}

	dbOperatorResults, err := h.OperatorResultRepo.GetByDAGResultBatch(
		ctx,
		[]uuid.UUID{args.dagResultID},
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving operator results.")
	}

	dbArtifactResults, err := h.ArtifactResultRepo.GetByDAGResults(
		ctx,
		[]uuid.UUID{args.dagResultID},
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving artifact results.")
	}

	dbArtifacts, err := h.ArtifactRepo.GetByDAG(ctx, dbDAG.ID, h.Database)
	dbArtifactsByID := functional_map.FromValues(
		dbArtifacts,
		func(artf models.Artifact) uuid.UUID { return artf.ID },
	)

	contents := make(map[string]string, len(dbArtifactResults))
	storageObj := storage.NewStorage(&dbDAG.StorageConfig)
	for _, artfResult := range dbArtifactResults {
		if artf, ok := dbArtifactsByID[artfResult.ArtifactID]; ok {
			// These artifacts has small content size and we can safely include them all in response.
			if artf.Type.IsCompact() {
				path := artfResult.ContentPath
				// Read data from storage and deserialize payload to `container`.
				contentBytes, err := storageObj.Get(ctx, path)
				if errors.Is(err, storage.ErrObjectDoesNotExist()) {
					// If the data does not exist, skip the fetch.
					continue
				}
				if err != nil {
					return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to get artifact content from storage")
				}
				contents[path] = string(contentBytes)
			}
		}
	}

	return response.NewNodeResultsFromDBObjects(
		dbOperatorResults, dbArtifactResults, contents,
	), http.StatusOK, nil
}
