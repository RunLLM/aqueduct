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
// Response: serialized `listIntegrationsResponse` containing all integrations accessible by the user.
//
// The caller must read the "exec_state" field on the result to determine if the integration was successfully
// registered.

type ListIntegrationsHandler struct {
	GetHandler

	Database database.Database

	IntegrationRepo repos.Integration
}

type listIntegrationsArgs struct {
	*aq_context.AqContext
}

type listIntegrationsResponse []integrationResponse

type integrationResponse struct {
	ID        uuid.UUID                `json:"id"`
	Service   shared.Service           `json:"service"`
	Name      string                   `json:"name"`
	Config    shared.IntegrationConfig `json:"config"`
	CreatedAt int64                    `json:"createdAt"`
	ExecState *shared.ExecutionState   `json:"exec_state"`
}

func (*ListIntegrationsHandler) Name() string {
	return "ListIntegrations"
}

func (*ListIntegrationsHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	return &listIntegrationsArgs{
		AqContext: aqContext,
	}, http.StatusOK, nil
}

func (h *ListIntegrationsHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*listIntegrationsArgs)

	emptyResponse := listIntegrationsResponse{}

	integrations, err := h.IntegrationRepo.GetByUser(
		ctx,
		args.OrgID,
		args.ID,
		h.Database,
	)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to list integrations.")
	}

	responses := make([]integrationResponse, 0, len(integrations))
	for _, integrationObject := range integrations {
		var response *integrationResponse
		var err error

		// If there is a Conda resource registered, embed additional configuration information inside Aqueduct Compute.
		// Otherwise, we simply note the current server's python version.
		if integrationObject.Name == shared.AqueductComputeName {
			var aqConfig shared.IntegrationConfig
			aqConfig, err = aqueduct_compute.ConstructAqueductComputeResourceConfig(ctx, args.ID, h.IntegrationRepo, h.Database)
			if err != nil {
				return emptyResponse, http.StatusInternalServerError, errors.Wrapf(err, "Unable to create aqueduct compute config!")
			}
			response, err = convertIntegrationObjectToResponse(&integrationObject, aqConfig)
		} else {
			response, err = convertIntegrationObjectToResponse(&integrationObject, integrationObject.Config)
		}
		if err != nil {
			return emptyResponse, http.StatusInternalServerError, errors.Wrapf(err, "Unable to create integration response for %s.", integrationObject.Name)
		}
		responses = append(responses, *response)
	}

	return responses, http.StatusOK, nil
}

// Helper function to convert an resource object into an integrationResponse
func convertIntegrationObjectToResponse(integrationObject *models.Integration, config shared.IntegrationConfig) (*integrationResponse, error) {
	execState, err := execution_state.ExtractConnectionState(integrationObject)
	if err != nil {
		return nil, err
	}

	return &integrationResponse{
		ID:        integrationObject.ID,
		Service:   integrationObject.Service,
		Name:      integrationObject.Name,
		Config:    config,
		CreatedAt: integrationObject.CreatedAt.Unix(),
		ExecState: execState,
	}, nil
}
