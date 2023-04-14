package v2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/response"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// This file should map directly to
// src/ui/common/src/handlers/v2/NodeArtifactResultsGet.tsx
//
// Route: /api/v2/workflow/{workflowID}/dag/{dagID}/node/artifact/{nodeID}/results
// Method: GET
// Params:
//	`workflowID`: ID for `workflow` object
//  `dagID`: ID for `workflow_dag` object
//	`nodeID`: ID for operator object
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		`[]response.ArtifactResult`

type NodeArtifactResultsGetHandler struct {
	nodeGetHandler
	handler.GetHandler

	Database database.Database

	WorkflowRepo       repos.Workflow
	DAGRepo            repos.DAG
	ArtifactRepo       repos.Artifact
	ArtifactResultRepo repos.ArtifactResult
}

func (*NodeArtifactResultsGetHandler) Name() string {
	return "NodeArtifactResultsGet"
}

func (h *NodeArtifactResultsGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
	return h.nodeGetHandler.Prepare(r)
}

func (h *NodeArtifactResultsGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*nodeGetArgs)

	artfID := args.nodeID
	wfID := args.workflowID

	emptyResponse := []response.ArtifactResult{}

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

	responses := make([]response.ArtifactResult, 0, len(results))
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

				responses = append(responses, *response.NewArtifactResultFromDBObject(
					&artfResult, contentPtr,
				))
			}
		} else {
			return emptyResponse, http.StatusInternalServerError, errors.Newf("Error retrieving dag %s", dbDagId)
		}
	}

	return responses, http.StatusOK, nil
}
