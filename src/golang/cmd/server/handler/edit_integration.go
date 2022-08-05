package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/request"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	postgres_utils "github.com/aqueducthq/aqueduct/lib/collections/utils"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ConnectIntegrationHandler connects a new integration for the organization.
type EditIntegrationHandler struct {
	PostHandler

	Database          database.Database
	IntegrationReader integration.Reader
	IntegrationWriter integration.Writer
	Vault             vault.Vault
	JobManager        job.JobManager
	StorageConfig     *shared.StorageConfig
}

func (*EditIntegrationHandler) Headers() []string {
	return []string{
		routes.IntegrationNameHeader,
		routes.IntegrationConfigHeader,
	}
}

type EditIntegrationArgs struct {
	*aq_context.AqContext
	Name          string
	IntegrationId uuid.UUID
	UpdatedFields map[string]string
}

type EditIntegrationResponse struct{}

func (*EditIntegrationHandler) Name() string {
	return "EditIntegration"
}

func (h *EditIntegrationHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, errors.Wrap(err, "Unable to edit integration.")
	}

	integrationIdStr := chi.URLParam(r, routes.IntegrationIdUrlParam)
	integrationId, err := uuid.Parse(integrationIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed integration ID.")
	}

	name, configMap, err := request.ParseIntegrationConfigFromRequest(r)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to edit integration.")
	}

	return &EditIntegrationArgs{
		AqContext:     aqContext,
		IntegrationId: integrationId,
		Name:          name,
		UpdatedFields: configMap,
	}, http.StatusOK, nil
}

func (h *EditIntegrationHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*EditIntegrationArgs)
	id := args.IntegrationId

	emptyResp := EditIntegrationResponse{}

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

	staticConfig, ok := config.()
	// Validate integration config
	statusCode, err := ValidateConfig(
		ctx,
		args.RequestId,
		args.Config,
		args.Service,
		h.JobManager,
		h.StorageConfig,
	)
	if err != nil {
		return emptyResp, statusCode, err
	}

	if statusCode, err := ConnectIntegration(ctx, args, h.IntegrationWriter, h.Database, h.Vault); err != nil {
		return emptyResp, statusCode, err
	}

	return emptyResp, http.StatusOK, nil
}

// ConnectIntegration connects a new integration specified by `args`. It returns a status code for the request
// and an error, if any.
func ConnectIntegration(
	ctx context.Context,
	args *ConnectIntegrationArgs,
	integrationWriter integration.Writer,
	db database.Database,
	vaultObject vault.Vault,
) (int, error) {
	// Extract non-confidential config
	publicConfig := args.Config.PublicConfig()

	var integrationObject *integration.Integration
	var err error
	if args.UserOnly {
		// This is a user-specific integration
		integrationObject, err = integrationWriter.CreateIntegrationForUser(
			ctx,
			args.OrganizationId,
			args.Id,
			args.Service,
			args.Name,
			(*postgres_utils.Config)(&publicConfig),
			true,
			db,
		)
	} else {
		integrationObject, err = integrationWriter.CreateIntegration(
			ctx,
			args.OrganizationId,
			args.Service,
			args.Name,
			(*postgres_utils.Config)(&publicConfig),
			true,
			db,
		)
	}
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to connect integration.")
	}

	// Store config (including confidential information) as in vault
	if err := auth.WriteConfigToSecret(
		ctx,
		integrationObject.Id,
		args.Config,
		vaultObject,
	); err != nil {
		// TODO ENG-498: Rollback integration write
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to connect integration.")
	}

	return http.StatusOK, nil
}
