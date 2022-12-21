package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Route: /integration/{integrationId}/test
// Method: POST
// Params: integrationId
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//
// Response: none, we expect caller to determine success / failure based on
// http status in addition to error message.
//
// TestIntegrationHandler tries to connect to an existing integration.
type TestIntegrationHandler struct {
	PostHandler

	Database   database.Database
	Vault      vault.Vault
	JobManager job.JobManager

	IntegrationRepo repos.Integration
}

type TestIntegrationArgs struct {
	*aq_context.AqContext
	IntegrationId uuid.UUID
}

type TestIntegrationResponse struct{}

func (*TestIntegrationHandler) Name() string {
	return "TestIntegration"
}

func (h *TestIntegrationHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	integrationIDStr := chi.URLParam(r, routes.IntegrationIdUrlParam)
	integrationID, err := uuid.Parse(integrationIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed integration ID.")
	}

	hasPermission, err := h.IntegrationRepo.ValidateOwnership(
		r.Context(),
		integrationID,
		aqContext.OrgID,
		aqContext.ID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Error validating integraiton ownership.")
	}

	if !hasPermission {
		return nil, http.StatusForbidden, errors.New("You don't have permission to access this integration.")
	}

	return &TestIntegrationArgs{AqContext: aqContext, IntegrationId: integrationID}, http.StatusOK, nil
}

func (h *TestIntegrationHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*TestIntegrationArgs)
	ID := args.IntegrationId

	emptyResp := TestIntegrationResponse{}

	integrationObject, err := h.IntegrationRepo.Get(ctx, ID, h.Database)
	if err == database.ErrNoRows {
		return emptyResp, http.StatusBadRequest, err
	}

	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to retrieve integration")
	}

	config, err := auth.ReadConfigFromSecret(ctx, ID, h.Vault)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to retrieve secrets")
	}

	// Validate integration config
	statusCode, err := ValidateConfig(
		ctx,
		args.RequestID,
		config,
		integrationObject.Service,
		h.JobManager,
		args.StorageConfig,
	)

	return emptyResp, statusCode, err
}
