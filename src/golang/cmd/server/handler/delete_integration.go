package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/config"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	aq_errors "github.com/aqueducthq/aqueduct/lib/errors"
	exec_env "github.com/aqueducthq/aqueduct/lib/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
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
	integrationObject            *models.Integration
	skipActiveWorkflowValidation bool
}

type deleteIntegrationResponse struct{}

type DeleteIntegrationHandler struct {
	PostHandler

	Database database.Database

	DAGRepo                  repos.DAG
	ExecutionEnvironmentRepo repos.ExecutionEnvironment
	IntegrationRepo          repos.Integration
	OperatorRepo             repos.Operator
	StorageMigrationRepo     repos.StorageMigration
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

	integrationObject, err := h.IntegrationRepo.Get(r.Context(), integrationID, h.Database)
	if err != nil {
		return nil, http.StatusNotFound, errors.Wrap(err, "Failed to retrieve integration object.")
	}

	if integrationObject.Service == shared.Kubernetes {
		if _, ok := integrationObject.Config[shared.K8sCloudIntegrationIdKey]; ok {
			return nil, http.StatusUnprocessableEntity, errors.Wrap(err, "Cannot delete the Aqueduct-generated k8s integration. Please delete the corresponding cloud integration instead.")
		}
	}

	// Built-in resources cannot be deleted.
	if shared.IsBuiltinResource(integrationObject.Name, integrationObject.Service) {
		return nil, http.StatusBadRequest, errors.New("Cannot delete built-in resources.")
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

	// Check that we can't delete an integration that is being used as artifact storage.
	currentStorageMigrationEntry, err := h.StorageMigrationRepo.Current(r.Context(), h.Database)
	if err != nil && !aq_errors.Is(err, database.ErrNoRows()) {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving current storage migration entry.")
	}
	if currentStorageMigrationEntry != nil && currentStorageMigrationEntry.DestIntegrationID == integrationObject.ID {
		return nil, http.StatusBadRequest, errors.New("Cannot delete an integration that is being used as artifact storage.")
	}

	return &deleteIntegrationArgs{
		AqContext:                    aqContext,
		integrationObject:            integrationObject,
		skipActiveWorkflowValidation: false,
	}, http.StatusOK, nil
}

func (h *DeleteIntegrationHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*deleteIntegrationArgs)
	emptyResp := deleteIntegrationResponse{}

	if !args.skipActiveWorkflowValidation {
		if statusCode, err := validateNoActiveWorkflowOnIntegration(
			ctx,
			args.AqContext,
			args.integrationObject,
			h.OperatorRepo,
			h.DAGRepo,
			h.IntegrationRepo,
			h.Database,
		); err != nil {
			return emptyResp, statusCode, err
		}
	}

	if args.integrationObject.Service == shared.AWS {
		// Note that this will make a call to DeleteIntegrationHandler.Perform() to delete the
		// Aqueduct-generated dynamic k8s integration.
		if statusCode, err := deleteCloudIntegrationHelper(ctx, args, h); err != nil {
			return emptyResp, statusCode, err
		}
	}

	txn, err := h.Database.BeginTx(ctx)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to delete integration.")
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	err = h.IntegrationRepo.Delete(ctx, args.integrationObject.ID, txn)
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
		args.integrationObject,
		h.OperatorRepo,
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
	aqContext *aq_context.AqContext,
	integrationObject *models.Integration,
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
	}).Perform(ctx, &listOperatorsForIntegrationArgs{AqContext: aqContext, integrationObject: integrationObject})
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
	operatorRepo repos.Operator,
	workflowRepo repos.Workflow,
	vaultObject vault.Vault,
	DB database.Database,
) error {
	if integrationObject.Service == shared.Conda {
		// Best effort to clean up
		err := exec_env.CleanupUnusedEnvironments(
			ctx, operatorRepo, DB,
		)
		if err != nil {
			return err
		}

		return exec_env.DeleteBaseEnvs()
	}

	if integrationObject.Service == shared.Email || integrationObject.Service == shared.Slack {
		err := workflowRepo.RemoveNotificationFromSettings(ctx, integrationObject.ID, DB)
		if err != nil {
			return err
		}
	}

	return vaultObject.Delete(ctx, integrationObject.ID.String())
}
