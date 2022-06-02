package server

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	postgres_utils "github.com/aqueducthq/aqueduct/lib/collections/utils"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type ListIntegrationsHandler struct {
	GetHandler

	Database          database.Database
	IntegrationReader integration.Reader
}

type listIntegrationsArgs struct {
	*aq_context.AqContext
}

type listIntegrationsResponse []integrationResponse

type integrationResponse struct {
	Id        uuid.UUID             `json:"id"`
	Service   integration.Service   `json:"service"`
	Name      string                `json:"name"`
	Config    postgres_utils.Config `json:"config"`
	CreatedAt int64                 `json:"createdAt"`
	Validated bool                  `json:"validated"`
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

	integrations, err := h.IntegrationReader.GetIntegrationsByUser(
		ctx,
		args.OrganizationId,
		args.Id,
		h.Database,
	)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to list integrations.")
	}

	responses := make([]integrationResponse, 0, len(integrations))
	for _, integrationObject := range integrations {
		response := convertIntegrationObjectToResponse(&integrationObject)
		responses = append(responses, *response)
	}

	return responses, http.StatusOK, nil
}

// Helper function to convert an Integration object into an integrationResponse
func convertIntegrationObjectToResponse(integrationObject *integration.Integration) *integrationResponse {
	return &integrationResponse{
		Id:        integrationObject.Id,
		Service:   integrationObject.Service,
		Name:      integrationObject.Name,
		Config:    integrationObject.Config,
		CreatedAt: integrationObject.CreatedAt.Unix(),
		Validated: integrationObject.Validated,
	}
}
