package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/lib/constants"
)

type getServerVersionResponse struct {
	Version string `json:"version"`
}

// Route: /api/version
// Method: GET
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//
// Response: Aqueduct server's current version number.
type GetServerVersionHandler struct {
	GetHandler
}

func (*GetServerVersionHandler) Name() string {
	return "GetServerVersion"
}

func (*GetServerVersionHandler) Prepare(r *http.Request) (interface{}, int, error) {
	return nil, http.StatusOK, nil
}

func (h *GetServerVersionHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	return getServerVersionResponse{
		Version: constants.ServerVersionNumber,
	}, http.StatusOK, nil
}
