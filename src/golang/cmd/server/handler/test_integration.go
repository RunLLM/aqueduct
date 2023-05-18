package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/config"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Route: /integration/{resourceID}/test
// Method: POST
// Params: resourceID
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//
// Response: none, we expect caller to determine success / failure based on
// http status in addition to error message.
//
// TestResourceHandler tries to connect to an existing integration.
type TestResourceHandler struct {
	PostHandler

	Database   database.Database
	JobManager job.JobManager

	ResourceRepo repos.Resource
}

type TestResourceArgs struct {
	*aq_context.AqContext
	ResourceId uuid.UUID
}

type TestResourceResponse struct{}

func (*TestResourceHandler) Name() string {
	return "TestIntegration"
}

func (h *TestResourceHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	resourceIDStr := chi.URLParam(r, routes.ResourceIDUrlParam)
	resourceID, err := uuid.Parse(resourceIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed resource ID.")
	}

	hasPermission, err := h.ResourceRepo.ValidateOwnership(
		r.Context(),
		resourceID,
		aqContext.OrgID,
		aqContext.ID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Error validating integraiton ownership.")
	}

	if !hasPermission {
		return nil, http.StatusForbidden, errors.New("You don't have permission to access this resource.")
	}

	return &TestResourceArgs{AqContext: aqContext, ResourceId: resourceID}, http.StatusOK, nil
}

func (h *TestResourceHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*TestResourceArgs)
	ID := args.ResourceId

	emptyResp := TestResourceResponse{}

	resourceObject, err := h.ResourceRepo.Get(ctx, ID, h.Database)
	if errors.Is(err, database.ErrNoRows()) {
		return emptyResp, http.StatusBadRequest, err
	}
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to retrieve resource")
	}

	// No need to do any further verification for Aqueduct Compute.
	// The fact that it even got here means it works.
	if resourceObject.Name == shared.AqueductComputeName {
		return emptyResp, http.StatusOK, nil
	}

	storageConfig := config.Storage()
	vaultObject, err := vault.NewVault(&storageConfig, config.EncryptionKey())
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to initialize vault.")
	}

	config, err := auth.ReadConfigFromSecret(ctx, ID, vaultObject)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to retrieve secrets")
	}

	// Validate resource config
	statusCode, err := ValidateConfig(
		ctx,
		args.RequestID,
		config,
		resourceObject.Service,
		h.JobManager,
		args.StorageConfig,
	)

	return emptyResp, statusCode, err
}
