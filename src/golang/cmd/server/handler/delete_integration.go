package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/queries"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	db_exec_env "github.com/aqueducthq/aqueduct/lib/collections/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	exec_env "github.com/aqueducthq/aqueduct/lib/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/repos"
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

	Database database.Database
	Vault    vault.Vault

	CustomReader      queries.Reader
	IntegrationReader integration.Reader
	// TODO: Replace with repos.Operator once ExecEnv methods are added
	OperatorReader operator.Reader

	IntegrationWriter          integration.Writer
	ExecutionEnvironmentReader db_exec_env.Reader
	ExecutionEnvironmentWriter db_exec_env.Writer

	OperatorRepo repos.Operator
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
		aqContext.OrgID,
		aqContext.ID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during integration ownership validation.")
	}

	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this integration.")
	}

	return &deleteIntegrationArgs{
		AqContext:     aqContext,
		integrationId: integrationId,
	}, http.StatusOK, nil
}

func (h *DeleteIntegrationHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*deleteIntegrationArgs)
	emptyResp := deleteIntegrationResponse{}

	code, err := validateNoActiveWorkflowOnIntegration(
		ctx,
		args.integrationId,
		h.OperatorReader,
		h.CustomReader,
		h.IntegrationReader,
		h.Database,
	)
	if err != nil {
		return emptyResp, code, err
	}

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

	if err := cleanUpIntegration(
		ctx,
		integrationObject,
		h.ExecutionEnvironmentReader,
		h.ExecutionEnvironmentWriter,
		h.Vault,
		txn,
	); err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to delete integration.")
	}

	if err := txn.Commit(ctx); err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to delete integration.")
	}

	return emptyResp, http.StatusOK, nil
}

// validateNoActiveWorkflowOnIntegration
// verifies there's no active workflow using the integration given the integration ID.
// It errors if there's any error occurred and passes if there's indeed no active workflow
// using that integration.
func validateNoActiveWorkflowOnIntegration(
	ctx context.Context,
	id uuid.UUID,
	operatorReader operator.Reader,
	customReader queries.Reader,
	integrationReader integration.Reader,
	db database.Database,
) (int, error) {
	interfaceResp, code, err := (&ListOperatorsForIntegrationHandler{
		CustomReader:      customReader,
		OperatorReader:    operatorReader,
		IntegrationReader: integrationReader,
		Database:          db,
	}).Perform(ctx, id)
	if err != nil {
		return code, errors.Wrap(err, "Error getting operators on this integration.")
	}

	operatorsOnIntegrationResp, ok := interfaceResp.(listOperatorsForIntegrationResponse)
	if !ok {
		return http.StatusInternalServerError, errors.New("Error getting operators on this integration.")
	}

	operatorsOnIntegration := operatorsOnIntegrationResp.OperatorWithIds
	for _, opState := range operatorsOnIntegration {
		if opState.IsActive {
			return http.StatusBadRequest, errors.New("We cannot delete this integration. There are still active workflows using it.")
		}
	}

	return http.StatusOK, nil
}

// cleanUpIntegration deletes any side effects of an integration
// in Aqueduct system, other than DB records.
// For example, credentials stored in vault or base conda environments
// created.
func cleanUpIntegration(
	ctx context.Context,
	integrationObject *integration.Integration,
	execEnvReader db_exec_env.Reader,
	execEnvWriter db_exec_env.Writer,
	vaultObject vault.Vault,
	db database.Database,
) error {
	if integrationObject.Service == integration.Conda {
		// Best effort to clean up
		err := exec_env.CleanupUnusedEnvironments(
			ctx, execEnvReader, execEnvWriter, db,
		)
		if err != nil {
			return err
		}

		return exec_env.DeleteBaseEnvs()
	}

	return vaultObject.Delete(ctx, integrationObject.Id.String())
}
