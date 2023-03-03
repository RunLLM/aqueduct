package handler

import (
	"context"
	"net/http"

	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	"github.com/aqueducthq/aqueduct/lib/repos"
)

// Route: /config/storage
// Method: POST
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//
// Response: none
type ConfigureStorageHandler struct {
	PostHandler

	Database database.Database
	Engine   engine.Engine

	ArtifactRepo repos.Artifact
	DAGRepo      repos.DAG
	DAGEdgeRepo  repos.DAGEdge
	OperatorRepo repos.Operator
	WorkflowRepo repos.Workflow
}

type configureStorageArgs struct {
	*aq_context.AqContext
}

func (*ConfigureStorageHandler) Name() string {
	return "ConfigureStorage"
}

func (h *ConfigureStorageHandler) Prepare(r *http.Request) (interface{}, int, error) {
	return nil, -1, nil
}

func (h *ConfigureStorageHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	return nil, -1, nil
}
