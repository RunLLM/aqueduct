package v2

import (
	"context"
	"net/http"
	"os"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/lib"
)

const inK8sClusterEnvVarName = "AQUEDUCT_IN_K8S_CLUSTER"

type getEnvironmentResponse struct {
	// Whether the server is running within a k8s cluster.
	InK8sCluster bool   `json:"inK8sCluster"`
	Version      string `json:"version"`
}

// Route: /api/v2/environment
// This file should map directly to
// src/ui/common/src/handlers/v2/EnvironmentGet.tsx
//
// Method: GET
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//
// Response: Aqueduct server's environment variables.
type EnvironmentHandler struct {
	handler.GetHandler
}

func (*EnvironmentHandler) Name() string {
	return "GetEnvironment"
}

func (*EnvironmentHandler) Prepare(r *http.Request) (interface{}, int, error) {
	return nil, http.StatusOK, nil
}

func (h *EnvironmentHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	inCluster := false
	if os.Getenv(inK8sClusterEnvVarName) == "1" {
		inCluster = true
	}

	return getEnvironmentResponse{
		InK8sCluster: inCluster,
		Version:      lib.ServerVersionNumber,
	}, http.StatusOK, nil
}
