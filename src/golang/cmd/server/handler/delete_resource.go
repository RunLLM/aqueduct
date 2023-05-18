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

// Route: /resource/{resourceID}/delete
// Method: POST
// Params:
//	`resourceID`: ID for `resource` object
// Request:
//	Headers:
//		`api-key`: user's API Key

// The `DeleteResourceHandler` does a best effort at deleting an resource.
type deleteResourceArgs struct {
	*aq_context.AqContext
	resourceObject               *models.Resource
	skipActiveWorkflowValidation bool
}

type deleteResourceResponse struct{}

type DeleteResourceHandler struct {
	PostHandler

	Database database.Database

	DAGRepo                  repos.DAG
	ExecutionEnvironmentRepo repos.ExecutionEnvironment
	ResourceRepo             repos.Resource
	OperatorRepo             repos.Operator
	StorageMigrationRepo     repos.StorageMigration
	WorkflowRepo             repos.Workflow
}

func (*DeleteResourceHandler) Name() string {
	return "DeleteResource"
}

func (h *DeleteResourceHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statuscode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statuscode, err
	}

	resourceIDStr := chi.URLParam(r, routes.ResourceIDUrlParam)
	resourceID, err := uuid.Parse(resourceIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed resource ID.")
	}

	resourceObject, err := h.ResourceRepo.Get(r.Context(), resourceID, h.Database)
	if err != nil {
		return nil, http.StatusNotFound, errors.Wrap(err, "Failed to retrieve resource object.")
	}

	if resourceObject.Service == shared.Kubernetes {
		if _, ok := resourceObject.Config[shared.K8sCloudResourceIdKey]; ok {
			return nil, http.StatusUnprocessableEntity, errors.Wrap(err, "Cannot delete the Aqueduct-generated k8s resource. Please delete the corresponding cloud resource instead.")
		}
	}

	// Built-in resources cannot be deleted.
	if shared.IsBuiltinResource(resourceObject.Name, resourceObject.Service) {
		return nil, http.StatusBadRequest, errors.New("Cannot delete built-in resources.")
	}

	ok, err := h.ResourceRepo.ValidateOwnership(
		r.Context(),
		resourceID,
		aqContext.OrgID,
		aqContext.ID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during resource ownership validation.")
	}

	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this resource.")
	}

	// Check that we can't delete an resource that is being used as artifact storage.
	currentStorageMigrationEntry, err := h.StorageMigrationRepo.Current(r.Context(), h.Database)
	if err != nil && !aq_errors.Is(err, database.ErrNoRows()) {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while retrieving current storage migration entry.")
	}
	if currentStorageMigrationEntry != nil && currentStorageMigrationEntry.DestResourceID == resourceObject.ID {
		return nil, http.StatusBadRequest, errors.New("Cannot delete an resource that is being used as artifact storage.")
	}

	return &deleteResourceArgs{
		AqContext:                    aqContext,
		resourceObject:               resourceObject,
		skipActiveWorkflowValidation: false,
	}, http.StatusOK, nil
}

func (h *DeleteResourceHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*deleteResourceArgs)
	emptyResp := deleteResourceResponse{}

	if !args.skipActiveWorkflowValidation {
		if statusCode, err := validateNoActiveWorkflowOnResource(
			ctx,
			args.AqContext,
			args.resourceObject,
			h.OperatorRepo,
			h.DAGRepo,
			h.ResourceRepo,
			h.Database,
		); err != nil {
			return emptyResp, statusCode, err
		}
	}

	if args.resourceObject.Service == shared.AWS {
		// Note that this will make a call to DeleteResourceHandler.Perform() to delete the
		// Aqueduct-generated dynamic k8s resource.
		if statusCode, err := deleteCloudResourceHelper(ctx, args, h); err != nil {
			return emptyResp, statusCode, err
		}
	}

	txn, err := h.Database.BeginTx(ctx)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to delete resource.")
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	err = h.ResourceRepo.Delete(ctx, args.resourceObject.ID, txn)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred while deleting resource.")
	}

	storageConfig := config.Storage()
	vaultObject, err := vault.NewVault(&storageConfig, config.EncryptionKey())
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to initialize vault.")
	}

	if err := cleanUpResource(
		ctx,
		args.resourceObject,
		h.OperatorRepo,
		h.WorkflowRepo,
		vaultObject,
		txn,
	); err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to delete resource.")
	}

	if err := txn.Commit(ctx); err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to delete resource.")
	}

	return emptyResp, http.StatusOK, nil
}

// validateNoActiveWorkflowOnResource
// verifies there's no active workflow using the resource given the resource ID.
// It errors if there's any error occurred and passes if there's indeed no active workflow
// using that resource.
func validateNoActiveWorkflowOnResource(
	ctx context.Context,
	aqContext *aq_context.AqContext,
	resourceObject *models.Resource,
	operatorRepo repos.Operator,
	dagRepo repos.DAG,
	resourceRepo repos.Resource,
	DB database.Database,
) (int, error) {
	interfaceResp, code, err := (&ListOperatorsResourecHandler{
		Database: DB,

		DAGRepo:      dagRepo,
		ResourceRepo: resourceRepo,
		OperatorRepo: operatorRepo,
	}).Perform(ctx, &listOperatorsForResourceArgs{AqContext: aqContext, resourceObject: resourceObject})
	if err != nil {
		return code, errors.Wrap(err, "Error getting operators on this resource.")
	}

	operatorsOnResourceResp, ok := interfaceResp.(listOperatorsForResourceResponse)
	if !ok {
		return http.StatusInternalServerError, errors.New("Error getting operators on this resource.")
	}

	operatorsOnResource := operatorsOnResourceResp.OperatorWithIds
	for _, opState := range operatorsOnResource {
		if opState.IsActive {
			return http.StatusBadRequest, errors.New("We cannot delete this resource. There are still active workflows using it.")
		}
	}

	return http.StatusOK, nil
}

// cleanUpResource deletes any side effects of an resource
// in Aqueduct system.
// For example, credentials stored in vault or base conda environments
// created.
func cleanUpResource(
	ctx context.Context,
	resourceObject *models.Resource,
	operatorRepo repos.Operator,
	workflowRepo repos.Workflow,
	vaultObject vault.Vault,
	DB database.Database,
) error {
	if resourceObject.Service == shared.Conda {
		// Best effort to clean up
		err := exec_env.CleanupUnusedEnvironments(
			ctx, operatorRepo, DB,
		)
		if err != nil {
			return err
		}

		return exec_env.DeleteBaseEnvs()
	}

	if resourceObject.Service == shared.Email || resourceObject.Service == shared.Slack {
		err := workflowRepo.RemoveNotificationFromSettings(ctx, resourceObject.ID, DB)
		if err != nil {
			return err
		}
	}

	return vaultObject.Delete(ctx, resourceObject.ID.String())
}
