package v2

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/response"
	"github.com/dropbox/godropbox/errors"
)

// This file should map directly to
// src/ui/common/src/handlers/v2/NodeMetricGet.tsx
//
// Route: /api/v2/workflow/{workflowID}/dag/{dagID}/node/metric/{nodeID}
// Method: GET
// Params:
//	`workflowID`: ID for `workflow` object
//  `dagID`: ID for `workflow_dag` object
//	`nodeID`: ID for metric operator object
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		`response.OperatorWithArtifactNode`

type NodeMetricGetHandler struct {
	nodeGetHandler
	handler.GetHandler

	Database database.Database

	WorkflowRepo repos.Workflow
	OperatorRepo repos.Operator
}

func (*NodeMetricGetHandler) Name() string {
	return "NodeMetricGet"
}

func (h *NodeMetricGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
	return h.nodeGetHandler.Prepare(r)
}

func (h *NodeMetricGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*nodeGetArgs)

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

	dbOperator, err := h.OperatorRepo.Get(ctx, args.nodeID, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error reading operator.")
	}
	if !dbOperator.Spec.IsMetric() {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Node ID does not belong to a metric operator.")
	}

	dbOperatorWithArtifactNode, err := h.OperatorRepo.GetOperatorWithArtifactNode(ctx, args.nodeID, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error reading metric node.")
	}

	return response.NewOperatorWithArtifactNodeFromDBObject(dbOperatorWithArtifactNode), http.StatusOK, nil
}
