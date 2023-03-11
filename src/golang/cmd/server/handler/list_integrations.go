package handler

import (
	"context"
	"net/http"

	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
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
	Validated bool                     `json:"validated"`
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
		response := convertIntegrationObjectToResponse(&integrationObject)
		responses = append(responses, *response)
	}

	return responses, http.StatusOK, nil
}

// Helper function to convert an Integration object into an integrationResponse
func convertIntegrationObjectToResponse(integrationObject *models.Integration) *integrationResponse {
	return &integrationResponse{
		ID:        integrationObject.ID,
		Service:   integrationObject.Service,
		Name:      integrationObject.Name,
		Config:    integrationObject.Config,
		CreatedAt: integrationObject.CreatedAt.Unix(),
		Validated: integrationObject.Validated,
	}
}
