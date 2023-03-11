package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/config"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
)

type getConfigArgs struct {
	*aq_context.AqContext
}

type getConfigResponse struct {
	AqPath              string                     `json:"aqPath"`
	RetentionJobPeriod  string                     `json:"retentionJobPeriod"`
	ApiKey              string                     `json:"apiKey"`
	StorageConfigPublic shared.StorageConfigPublic `json:"storageConfig"`
}

type GetConfigHandler struct {
	GetHandler
}

func (*GetConfigHandler) Name() string {
	return "GetConfig"
}

func (h *GetConfigHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	return &getConfigArgs{
		AqContext: aqContext,
	}, http.StatusOK, nil
}

func (h *GetConfigHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	storageConfig := config.Storage()
	storageConfigPtr := &storageConfig
	storageConfigPublic, err := storageConfigPtr.ToPublic()
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve storage config.")
	}

	return getConfigResponse{
		AqPath:              config.AqueductPath(),
		RetentionJobPeriod:  config.RetentionJobPeriod(),
		ApiKey:              config.APIKey(),
		StorageConfigPublic: *storageConfigPublic,
	}, http.StatusOK, nil
}
