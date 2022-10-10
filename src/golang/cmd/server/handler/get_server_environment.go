package handler

import (
	"context"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

const inK8sClusterEnvVarName = "AQUEDUCT_IN_K8S_CLUSTER"

type getServerEnvironmentResponse struct {
	// Whether the server is running within a k8s cluster.
	InK8sCluster bool `json:"inK8sCluster"`
}

// Route: /api/environment
// Method: GET
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//
// Response: Aqueduct server's environment variables.
type GetServerEnvironmentHandler struct {
	GetHandler
}

func (*GetServerEnvironmentHandler) Name() string {
	return "GetServerEnvironment"
}

func (*GetServerEnvironmentHandler) Prepare(r *http.Request) (interface{}, int, error) {
	return nil, http.StatusOK, nil
}

func (h *GetServerEnvironmentHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	inCluster := false
	if os.Getenv(inK8sClusterEnvVarName) == "1" {
		inCluster = true
	}

	log.Info("Incluster is %s", inCluster)

	return getServerEnvironmentResponse{
		InK8sCluster: inCluster,
	}, http.StatusOK, nil
}
