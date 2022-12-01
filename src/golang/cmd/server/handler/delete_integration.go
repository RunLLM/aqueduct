package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	exec_env "github.com/aqueducthq/aqueduct/lib/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Route: /integration/{integrationId}/delete
// Method: POST
// Params:
//	`integrationId`: ID for `integration` object
// Request:
//	Headers:
//		`api-key`: user's API Key

// The `DeleteIntegrationHandler` does a best effort at deleting an integration.
type deleteIntegrationArgs struct {
	*aq_context.AqContext
	integrationId uuid.UUID
}

type deleteIntegrationResponse struct{}

type DeleteIntegrationHandler struct {
	PostHandler

	Database          database.Database
	Vault             vault.Vault
	OperatorReader    operator.Reader
	IntegrationReader integration.Reader
	IntegrationWriter integration.Writer
}

func (*DeleteIntegrationHandler) Name() string {
	return "DeleteIntegration"
}

func (h *DeleteIntegrationHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statuscode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statuscode, err
	}

	integrationIdStr := chi.URLParam(r, routes.IntegrationIdUrlParam)
	integrationId, err := uuid.Parse(integrationIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed integration ID.")
	}

	ok, err := h.IntegrationReader.ValidateIntegrationOwnership(
		r.Context(),
		integrationId,
		aqContext.OrganizationId,
		aqContext.Id,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during integration ownership validation.")
	}

	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this integration.")
	}

	// Fetch all operators on this integration.
	operators, err := h.OperatorReader.GetOperatorsByIntegrationId(r.Context(), integrationId, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve operators.")
	}
	if len(operators) > 0 {
		return nil, http.StatusBadRequest, errors.New("Unable to delete because the integration is in use.")
	}

	return &deleteIntegrationArgs{
		AqContext:     aqContext,
		integrationId: integrationId,
	}, http.StatusOK, nil
}

func (h *DeleteIntegrationHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*deleteIntegrationArgs)
	emptyResp := deleteIntegrationResponse{}

	integrationObject, err := h.IntegrationReader.GetIntegration(ctx, args.integrationId, h.Database)
	if err != nil {
		return emptyResp, http.StatusBadRequest, errors.Wrap(err, "failed to retrieve the given integration.")
	}

	txn, err := h.Database.BeginTx(ctx)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to delete integration.")
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	err = h.IntegrationWriter.DeleteIntegration(ctx, args.integrationId, txn)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting integration.")
	}

	if err := cleanUpIntegration(ctx, integrationObject, h.Vault); err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to delete integration.")
	}

	if err := txn.Commit(ctx); err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to delete integration.")
	}

	return emptyResp, http.StatusOK, nil
}

// cleanUpIntegration deletes any side effects of an integration
// in Aqueduct system, other than DB records.
// For example, credentials stored in vault or base conda environments
// created.
func cleanUpIntegration(
	ctx context.Context,
	integrationObject *integration.Integration,
	vaultObject vault.Vault,
) error {
	err := vaultObject.Delete(ctx, integrationObject.Id.String())
	if err != nil {
		return err
	}

	if integrationObject.Service == integration.Conda {
		return exec_env.DeleteBaseEnvs()
	}

	return nil
}
