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
// src/ui/common/src/handlers/v2/NodeArtifactGet.tsx
//
// Route: /api/v2/workflow/{workflowID}/dag/{dagID}/node/artifact/{nodeID}
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
//		`response.Artifact`

type NodeArtifactGetHandler struct {
	nodeGetHandler
	handler.GetHandler

	Database database.Database

	WorkflowRepo repos.Workflow
	ArtifactRepo repos.Artifact
}

func (*NodeArtifactGetHandler) Name() string {
	return "NodeArtifactGet"
}

func (h *NodeArtifactGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
	return h.nodeGetHandler.Prepare(r)
}

func (h *NodeArtifactGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
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

	dbArtifactNode, err := h.ArtifactRepo.GetNode(ctx, args.nodeID, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error reading artifact node.")
	}

	return response.NewArtifactFromDBObject(dbArtifactNode), http.StatusOK, nil
}
