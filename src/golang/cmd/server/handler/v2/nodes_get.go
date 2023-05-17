package v2

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/cmd/server/request/parser"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/response"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// This file should map directly to
// src/ui/common/src/handlers/v2/NodesGet.tsx
//
// Route: /v2/workflow/{workflowID}/dag/{dagID}/nodes
// Method: GET
// Params:
//	`workflowID`: ID for `workflow` object
//  `dagID`: ID for `workflow_dag` object
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		serialized `node.Nodes`

type nodesGetArgs struct {
	*aq_context.AqContext
	workflowID uuid.UUID
	dagID      uuid.UUID
}

type NodesGetHandler struct {
	handler.GetHandler

	Database database.Database

	WorkflowRepo repos.Workflow
	OperatorRepo repos.Operator
	ArtifactRepo repos.Artifact
}

func (*NodesGetHandler) Name() string {
	return "NodesGet"
}

func (h *NodesGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	workflowID, err := (parser.WorkflowIDParser{}).Parse(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	dagID, err := (parser.DagIDParser{}).Parse(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return &nodesGetArgs{
		AqContext:  aqContext,
		workflowID: workflowID,
		dagID:      dagID,
	}, http.StatusOK, nil
}

func (h *NodesGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*nodesGetArgs)

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

	dbOperatorNodes, err := h.OperatorRepo.GetNodesByDAG(ctx, args.dagID, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error reading operator nodes.")
	}

	dbArtifactNodes, err := h.ArtifactRepo.GetNodesByDAG(ctx, args.dagID, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error reading artifact nodes.")
	}

	// TODO: ENG-2987 Create separate sections for Metrics/Checks
	// dbMetricNodes, err := h.OperatorRepo.GetOperatorWithArtifactNodesByDAG(ctx, args.dagID, h.Database)
	// if err != nil {
	// 	return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error reading metric nodes.")
	// }

	// dbCheckNodes, err := h.OperatorRepo.GetOperatorWithArtifactNodesByDAG(ctx, args.dagID, h.Database)
	// if err != nil {
	// 	return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error reading check nodes.")
	// }

	return response.NewNodesFromDBObjects(
		dbOperatorNodes,
		dbArtifactNodes,
		// TODO: ENG-2987 Create separate sections for Metrics/Checks
		// dbMetricNodes,
		// dbCheckNodes,
	), http.StatusOK, nil
}
