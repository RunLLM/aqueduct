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

// Route: /api/integration/dynamic-engine/status
// Method: GET
// Params: None
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//		`integration_ids`: json serialized list of dynamic engine integration IDs
//
// Response: serialized `getDynamicEngineStatusResponse` which contains one entry per dynamic engine.
type GetDynamicEngineStatusHandler struct {
	GetHandler

	Database database.Database

	ResourceRepo repos.Resource
}

type getDynamicEngineStatusArgs struct {
	*aq_context.AqContext
	resourceIds []uuid.UUID
}

type getDynamicEngineStatusResponse []dynamicEngineStatusResponse

type dynamicEngineStatusResponse struct {
	ID     uuid.UUID                   `json:"id"`
	Name   string                      `json:"name"`
	Status shared.K8sClusterStatusType `json:"status"`
}

func (*GetDynamicEngineStatusHandler) Name() string {
	return "GetDynamicEngineStatus"
}

func (*GetDynamicEngineStatusHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	var resourceIdsStr []string
	resourceIdsJson := r.Header.Get(routes.IntegrationIDsHeader)
	err = json.Unmarshal([]byte(resourceIdsJson), &resourceIdsStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Error unmarshalling resource IDs.")
	}

	resourceIds := make([]uuid.UUID, 0, len(resourceIdsStr))
	for _, resourceIdStr := range resourceIdsStr {
		resourceId, err := uuid.Parse(resourceIdStr)
		if err != nil {
			return nil, http.StatusBadRequest, errors.Wrap(err, "Error parsing resource ID.")
		}
		resourceIds = append(resourceIds, resourceId)
	}

	return &getDynamicEngineStatusArgs{
		AqContext:   aqContext,
		resourceIds: resourceIds,
	}, http.StatusOK, nil
}

func (h *GetDynamicEngineStatusHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*getDynamicEngineStatusArgs)

	emptyResponse := getDynamicEngineStatusResponse{}

	resources, err := h.ResourceRepo.GetBatch(
		ctx,
		args.resourceIds,
		h.Database,
	)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to get engine resources.")
	}

	responses := make([]dynamicEngineStatusResponse, 0, len(resources))
	for _, resourceObject := range resources {
		if _, ok := resourceObject.Config[shared.K8sDynamicKey]; ok {
			if resourceObject.Config[shared.K8sDynamicKey] == strconv.FormatBool(true) {
				response := dynamicEngineStatusResponse{
					ID:     resourceObject.ID,
					Name:   resourceObject.Name,
					Status: shared.K8sClusterStatusType(resourceObject.Config[shared.K8sStatusKey]),
				}
				responses = append(responses, response)
			}
		}
	}

	return responses, http.StatusOK, nil
}
