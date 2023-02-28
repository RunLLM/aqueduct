package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// Route: /api/integration/engine/status
// Method: GET
// Params: None
// Request:
//
//		Headers:
//			`api-key`: user's API Key
//	     	`integration_ids` json serialized list of engine integration IDs
//
// Response: serialized `GetEngineStatusResponse`.
type GetEngineStatusHandler struct {
	GetHandler

	Database database.Database

	IntegrationRepo repos.Integration
}

type getEngineStatusArgs struct {
	*aq_context.AqContext
	integrationIds []uuid.UUID
}

type getEngineStatusResponse []engineStatusResponse

type engineStatusResponse struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Status string    `json:"status"`
}

func (*GetEngineStatusHandler) Name() string {
	return "GetEngineStatus"
}

func (*GetEngineStatusHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	var integrationIdsStr []string
	integrationIdsJson := r.Header.Get(routes.IntegrationIDsHeader)
	err = json.Unmarshal([]byte(integrationIdsJson), &integrationIdsStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Error unmarshalling integration IDs.")
	}

	integrationIds := make([]uuid.UUID, 0, len(integrationIdsStr))
	for _, integrationIdStr := range integrationIdsStr {
		integrationId, err := uuid.Parse(integrationIdStr)
		if err != nil {
			return nil, http.StatusBadRequest, errors.Wrap(err, "Error parsing integration ID.")
		}
		integrationIds = append(integrationIds, integrationId)
	}

	return &getEngineStatusArgs{
		AqContext:      aqContext,
		integrationIds: integrationIds,
	}, http.StatusOK, nil
}

func (h *GetEngineStatusHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*getEngineStatusArgs)

	emptyResponse := getEngineStatusResponse{}

	integrations, err := h.IntegrationRepo.GetBatch(
		ctx,
		args.integrationIds,
		h.Database,
	)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to get engine integrations.")
	}

	responses := make([]engineStatusResponse, 0, len(integrations))
	for _, integrationObject := range integrations {
		if _, ok := integrationObject.Config[shared.K8sDynamicKey]; ok {
			if integrationObject.Config[shared.K8sDynamicKey] == strconv.FormatBool(true) {
				response := engineStatusResponse{
					ID:     integrationObject.ID,
					Name:   integrationObject.Name,
					Status: integrationObject.Config[shared.K8sStatusKey],
				}
				responses = append(responses, response)
			}
		}
	}

	return responses, http.StatusOK, nil
}
