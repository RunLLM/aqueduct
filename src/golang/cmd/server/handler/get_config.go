package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/config"
)

type GetConfigHandler struct {
	GetHandler
}

func (*GetConfigHandler) Name() string {
	return "GetConfig"
}

func (h *GetConfigHandler) Prepare(r *http.Request) (interface{}, int, error) {
	globalConfig := config.GetGlobalConfig()
	return globalConfig, http.StatusOK, nil
}

func (h *GetConfigHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	globalConfig := config.GetGlobalConfig()
	return globalConfig, http.StatusOK, nil
}
