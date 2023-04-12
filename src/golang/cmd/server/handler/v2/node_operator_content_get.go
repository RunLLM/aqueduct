package v2

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	server_resp "github.com/aqueducthq/aqueduct/cmd/server/response"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/response"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/dropbox/godropbox/errors"
)

// This file should map directly to
// src/ui/common/src/handlers/NodeOperatorContentGet.tsx
//
// Route: /api/v2/workflow/{workflowID}/dag/{dagID}/node/operator/{nodeID}/content
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
//		`response.Content`

type NodeOperatorContentGetHandler struct {
	nodeGetHandler
	handler.GetHandler

	Database database.Database

	WorkflowRepo repos.Workflow
	DAGRepo      repos.DAG
	OperatorRepo repos.Operator
}

func (*NodeOperatorContentGetHandler) Name() string {
	return "NodeOperatorContentGet"
}

func (h *NodeOperatorContentGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
	return h.nodeGetHandler.Prepare(r)
}

func (h *NodeOperatorContentGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*nodeGetArgs)
	emptyResp := &response.Content{}

	ok, err := h.WorkflowRepo.ValidateOrg(
		ctx,
		args.workflowID,
		args.OrgID,
		h.Database,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during workflow ownership validation.")
	}

	if !ok {
		return emptyResp, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this workflow.")
	}

	dbOperator, err := h.OperatorRepo.Get(ctx, args.nodeID, h.Database)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error reading operator node.")
	}

	if !dbOperator.Spec.HasFunction() {
		return emptyResp, http.StatusOK, nil
	}

	dbDAG, err := h.DAGRepo.Get(ctx, args.dagID, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error reading DAG.")
	}

	if dbOperator.Spec.Function() == nil {
		return emptyResp, http.StatusInternalServerError, errors.New("Requested operator does not have function.")
	}

	path := dbOperator.Spec.Function().StoragePath
	storageObj := storage.NewStorage(&dbDAG.StorageConfig)
	content, err := storageObj.Get(ctx, path)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err,
			fmt.Sprintf("Error getting operator content %s", path),
		)
	}

	return &response.Content{Name: dbOperator.Name, Data: content}, http.StatusOK, nil
}

func (*NodeOperatorContentGetHandler) SendResponse(w http.ResponseWriter, interfaceResp interface{}) {
	resp := interfaceResp.(*response.Content)
	server_resp.SendSmallFileResponse(
		w,
		fmt.Sprintf("%s.zip", resp.Name),
		bytes.NewBuffer(resp.Data),
	)
}
