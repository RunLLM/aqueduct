package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// TestIntegrationHandler tries to connect to an existing integration.
type TestIntegrationHandler struct {
	PostHandler

	Database          database.Database
	IntegrationReader integration.Reader
	Vault             vault.Vault
	JobManager        job.JobManager
	StorageConfig     *shared.StorageConfig
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

	integrationIdStr := chi.URLParam(r, routes.IntegrationIdUrlParam)
	integrationId, err := uuid.Parse(integrationIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed integration ID.")
	}

	return &TestIntegrationArgs{AqContext: aqContext, IntegrationId: integrationId}, http.StatusOK, nil
}

func (h *TestIntegrationHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*TestIntegrationArgs)
	id := args.IntegrationId

	emptyResp := TestIntegrationResponse{}

	integrationObject, err := h.IntegrationReader.GetIntegration(ctx, id, h.Database)
	if err == database.ErrNoRows {
		return emptyResp, http.StatusBadRequest, err
	}

	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to retrieve integration")
	}

	config, err := auth.ReadConfigFromSecret(ctx, id, h.Vault)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to retrieve secrets")
	}

	// Validate integration config
	statusCode, err := ValidateConfig(
		ctx,
		args.RequestId,
		config,
		integrationObject.Service,
		h.JobManager,
		h.StorageConfig,
	)

	return emptyResp, statusCode, err
}
