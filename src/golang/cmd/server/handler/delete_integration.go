package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	exec_env "github.com/aqueducthq/aqueduct/lib/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/models"
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
	integrationID uuid.UUID
}

type deleteIntegrationResponse struct{}

type DeleteIntegrationHandler struct {
	PostHandler

	Database database.Database

	DAGRepo                  repos.DAG
	ExecutionEnvironmentRepo repos.ExecutionEnvironment
	IntegrationRepo          repos.Integration
	OperatorRepo             repos.Operator
	WorkflowRepo             repos.Workflow
}

func (*DeleteIntegrationHandler) Name() string {
	return "DeleteIntegration"
}

func (h *DeleteIntegrationHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statuscode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statuscode, err
	}

	integrationIDStr := chi.URLParam(r, routes.IntegrationIdUrlParam)
	integrationID, err := uuid.Parse(integrationIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed integration ID.")
	}

	ok, err := h.IntegrationRepo.ValidateOwnership(
		r.Context(),
		integrationID,
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
		integrationID: integrationID,
	}, http.StatusOK, nil
}

func (h *DeleteIntegrationHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*deleteIntegrationArgs)
	emptyResp := deleteIntegrationResponse{}

	code, err := validateNoActiveWorkflowOnIntegration(
		ctx,
		args.integrationID,
		h.OperatorRepo,
		h.DAGRepo,
		h.IntegrationRepo,
		h.Database,
	)
	if err != nil {
		return emptyResp, code, err
	}

	integrationObject, err := h.IntegrationRepo.Get(ctx, args.integrationID, h.Database)
	if err != nil {
		return emptyResp, http.StatusBadRequest, errors.Wrap(err, "failed to retrieve the given integration.")
	}

	txn, err := h.Database.BeginTx(ctx)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to delete integration.")
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	err = h.IntegrationRepo.Delete(ctx, args.integrationID, txn)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting integration.")
	}

	storageConfig := config.Storage()
	vaultObject, err := vault.NewVault(&storageConfig, config.EncryptionKey())
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to initialize vault.")
	}

	if err := cleanUpIntegration(
		ctx,
		integrationObject,
		h.ExecutionEnvironmentRepo,
		h.WorkflowRepo,
		vaultObject,
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
	operatorRepo repos.Operator,
	dagRepo repos.DAG,
	integrationRepo repos.Integration,
	DB database.Database,
) (int, error) {
	interfaceResp, code, err := (&ListOperatorsForIntegrationHandler{
		Database: DB,

		DAGRepo:         dagRepo,
		IntegrationRepo: integrationRepo,
		OperatorRepo:    operatorRepo,
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
// in Aqueduct system.
// For example, credentials stored in vault or base conda environments
// created.
func cleanUpIntegration(
	ctx context.Context,
	integrationObject *models.Integration,
	execEnvRepo repos.ExecutionEnvironment,
	workflowRepo repos.Workflow,
	vaultObject vault.Vault,
	DB database.Database,
) error {
	if integrationObject.Service == integration.Conda {
		// Best effort to clean up
		err := exec_env.CleanupUnusedEnvironments(
			ctx, execEnvRepo, DB,
		)
		if err != nil {
			return err
		}

		return exec_env.DeleteBaseEnvs()
	}

	if integrationObject.Service == integration.Email || integrationObject.Service == integration.Slack {
		err := workflowRepo.RemoveNotificationFromSettings(ctx, integrationObject.ID, DB)
		if err != nil {
			return err
		}
	}

	return vaultObject.Delete(ctx, integrationObject.ID.String())
}
