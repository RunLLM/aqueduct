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
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/google/uuid"
)

// This file should map directly to
// src/ui/common/src/handlers/v2/NodeMetricResultsGet.tsx
//
// Returns all downstream artifact results
// Route: /api/v2/workflow/{workflowID}/dag/{dagID}/node/metric/{nodeID}/results
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
//		`[]response.OperatorWithArtifactResultNode`

type NodeMetricResultsGetHandler struct {
	nodeGetHandler
	handler.GetHandler

	Database database.Database

	WorkflowRepo       repos.Workflow
	DAGRepo            repos.DAG
	OperatorRepo       repos.Operator
	OperatorResultRepo repos.OperatorResult
}

func (*NodeMetricResultsGetHandler) Name() string {
	return "NodeMetricResultsGet"
}

func (h *NodeMetricResultsGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
	return h.nodeGetHandler.Prepare(r)
}

func (h *NodeMetricResultsGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*nodeGetArgs)

	artfID := args.nodeID
	wfID := args.workflowID

	emptyResponse := []response.OperatorWithArtifactResultNode{}

	dbOperatorWithArtifactNodes, err := h.OperatorRepo.GetOperatorWithArtifactByArtifactIdNodeBatch(ctx, []uuid.UUID{artfID}, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error reading metric node.")
	}
	dbOperatorWithArtifactNode := views.OperatorWithArtifactNode{}
	if len(dbOperatorWithArtifactNodes) == 0 {
		return emptyResponse, http.StatusOK, nil
	} else {
		dbOperatorWithArtifactNode = dbOperatorWithArtifactNodes[0]
	}

	results, err := h.OperatorResultRepo.GetOperatorWithArtifactResultNodesByOperatorNameAndWorkflow(ctx, dbOperatorWithArtifactNode.Name, wfID, h.Database)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve metric results.")
	}

	if len(results) == 0 {
		return emptyResponse, http.StatusOK, nil
	}

	resultArtifactIds := make([]uuid.UUID, 0, len(results))
	for _, result := range results {
		resultArtifactIds = append(resultArtifactIds, result.ArtifactResultID)
	}

	artfResultToDAG, err := h.DAGRepo.GetByArtifactResultBatch(ctx, resultArtifactIds, h.Database)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve workflow dags.")
	}

	// maps from db dag Ids
	dbDagByDagId := make(map[uuid.UUID]models.DAG, len(artfResultToDAG))
	nodeResultByDagId := make(map[uuid.UUID][]views.OperatorWithArtifactResultNode, len(artfResultToDAG))
	for _, NodeResult := range results {
		if dbDag, ok := artfResultToDAG[NodeResult.ArtifactResultID]; ok {
			if _, okDagsMap := dbDagByDagId[dbDag.ID]; !okDagsMap {
				dbDagByDagId[dbDag.ID] = dbDag
			}

			nodeResultByDagId[dbDag.ID] = append(nodeResultByDagId[dbDag.ID], NodeResult)
		} else {
			return emptyResponse, http.StatusInternalServerError, errors.Newf("Error retrieving dag associated with artifact result %s", NodeResult.ArtifactResultID)
		}
	}

	responses := make([]response.OperatorWithArtifactResultNode, 0, len(results))
	for dbDagId, nodeResults := range nodeResultByDagId {
		if dag, ok := dbDagByDagId[dbDagId]; ok {
			storageObj := storage.NewStorage(&dag.StorageConfig)
			if err != nil {
				return emptyResponse, http.StatusInternalServerError, errors.New("Error retrieving artifact contents.")
			}

			for _, nodeResult := range nodeResults {
				var contentPtr *string = nil
				if !nodeResult.ArtifactResultExecState.IsNull &&
					(nodeResult.ArtifactResultExecState.ExecutionState.Status == shared.FailedExecutionStatus ||
					 nodeResult.ArtifactResultExecState.ExecutionState.Status == shared.SucceededExecutionStatus) {
					exists := storageObj.Exists(ctx, nodeResult.ContentPath)
					if exists {
						contentBytes, err := storageObj.Get(ctx, nodeResult.ContentPath)
						if err != nil {
							return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, fmt.Sprintf("Error retrieving artifact content for result %s", nodeResult.ArtifactID))
						}

						contentStr := string(contentBytes)
						contentPtr = &contentStr
					}
				}

				responses = append(responses, *response.NewOperatorWithArtifactResultNodeFromDBObject(
					&nodeResult, contentPtr,
				))
			}
		} else {
			return emptyResponse, http.StatusInternalServerError, errors.Newf("Error retrieving dag %s", dbDagId)
		}
	}

	return responses, http.StatusOK, nil
}
