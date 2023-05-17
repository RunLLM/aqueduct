package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/lib/aqueduct_compute"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/execution_state"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// Route: /integrations
// Method: GET
// Params: None
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//
// Response: serialized `listResourceResponse` containing all integrations accessible by the user.
//
// The caller must read the "exec_state" field on the result to determine if the integration was successfully
// registered.

type ListResourcesHandler struct {
	GetHandler

	Database database.Database

	ResourceRepo repos.Resource
}

type listResourceArgs struct {
	*aq_context.AqContext
}

type listResourceResponse []resourceResponse

type resourceResponse struct {
	ID        uuid.UUID              `json:"id"`
	Service   shared.Service         `json:"service"`
	Name      string                 `json:"name"`
	Config    shared.ResourceConfig  `json:"config"`
	CreatedAt int64                  `json:"createdAt"`
	ExecState *shared.ExecutionState `json:"exec_state"`
}

func (*ListResourcesHandler) Name() string {
	return "ListIntegrations"
}

func (*ListResourcesHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	return &listResourceArgs{
		AqContext: aqContext,
	}, http.StatusOK, nil
}

func (h *ListResourcesHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*listResourceArgs)

	emptyResponse := listResourceResponse{}

	resources, err := h.ResourceRepo.GetByUser(
		ctx,
		args.OrgID,
		args.ID,
		h.Database,
	)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to list resources.")
	}

	responses := make([]resourceResponse, 0, len(resources))
	for _, resourceObject := range resources {
		var response *resourceResponse
		var err error

		// If there is a Conda resource registered, embed additional configuration information inside Aqueduct Compute.
		// Otherwise, we simply note the current server's python version.
		if resourceObject.Name == shared.AqueductComputeName {
			var aqConfig shared.ResourceConfig
			aqConfig, err = aqueduct_compute.ConstructAqueductComputeResourceConfig(ctx, args.ID, h.ResourceRepo, h.Database)
			if err != nil {
				return emptyResponse, http.StatusInternalServerError, errors.Wrapf(err, "Unable to create aqueduct compute config!")
			}
			response, err = convertResourceObjectToResponse(&resourceObject, aqConfig)
		} else {
			response, err = convertResourceObjectToResponse(&resourceObject, resourceObject.Config)
		}
		if err != nil {
			return emptyResponse, http.StatusInternalServerError, errors.Wrapf(err, "Unable to create resource response for %s.", resourceObject.Name)
		}
		responses = append(responses, *response)
	}

	return responses, http.StatusOK, nil
}

// Helper function to convert an resource object into an resourceResponse
func convertResourceObjectToResponse(resourceObject *models.Resource, config shared.ResourceConfig) (*resourceResponse, error) {
	execState, err := execution_state.ExtractConnectionState(resourceObject)
	if err != nil {
		return nil, err
	}

	return &resourceResponse{
		ID:        resourceObject.ID,
		Service:   resourceObject.Service,
		Name:      resourceObject.Name,
		Config:    config,
		CreatedAt: resourceObject.CreatedAt.Unix(),
		ExecState: execState,
	}, nil
}
