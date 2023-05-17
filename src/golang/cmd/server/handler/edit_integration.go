package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/request"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/config"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Route: /integration/{resourceID}/edit
// Method: POST
// Params: resourceID
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//		`integration-name`: the updated name for the integration. Empty if no updates.
//		`integration-config`: the json-serialized integration config.
//							Could contain only updated fields. This field
//							can be empty if there's no config updates.
//
// Response: none
type EditResourceHandler struct {
	PostHandler

	Database   database.Database
	JobManager job.JobManager

	ResourceRepo repos.Resource
}

var serviceToReadOnlyFields = map[shared.Service]map[string]bool{
	shared.Airflow:  {"host": true},
	shared.BigQuery: {"project_id": true},
	shared.MariaDb: {
		"host":     true,
		"port":     true,
		"database": true,
	},
	shared.MySql: {
		"host":     true,
		"port":     true,
		"database": true,
	},
	shared.Postgres: {
		"host":     true,
		"port":     true,
		"database": true,
	},
	shared.Redshift: {
		"host":     true,
		"port":     true,
		"database": true,
	},
	shared.S3: {
		"bucket":         true,
		"region":         true,
		"use_as_storage": true,
	},
	shared.Snowflake: {
		"account_identifier": true,
		"warehouse":          true,
		"database":           true,
	},
}

func (*EditResourceHandler) Headers() []string {
	return []string{
		routes.IntegrationNameHeader,
		routes.IntegrationConfigHeader,
	}
}

type EditResourceArgs struct {
	*aq_context.AqContext
	Name          string
	ResourceID    uuid.UUID
	UpdatedFields map[string]string
}

type EditResourceResponse struct{}

func (*EditResourceHandler) Name() string {
	return "EditIntegration"
}

// `updateConfig` updates `curConfigToUpdate` *in-place* with `newConfig` with
// the same behavior as map updates.
// It returns 3 values:
// - whether there's actually an update
// - http status code
// - error if there's any
//
// If trying to update a 'read only' field defined by `ServerToReadOnlyFieldsMap`,
// `updateConfig` will return a 400 and an error.
func updateConfig(
	curConfigToUpdate map[string]string,
	service shared.Service,
	newConfig map[string]string,
) (bool, int, error) {
	readOnlyFields := serviceToReadOnlyFields[service]
	updated := false
	for k, v := range newConfig {
		if v == "" {
			continue // no update occurs
		}

		_, isReadonlyField := readOnlyFields[k]
		curValue, existsInCurConfig := curConfigToUpdate[k]
		if isReadonlyField && existsInCurConfig && curValue != v {
			// Throw if:
			// * field is read-only, and
			// * field both exists in cur and new, and
			// * field values are different in cur and new
			return false, http.StatusBadRequest, errors.Newf(
				"Error updating read-only field %s. For %s, %v are read-only fields which cannot be edited.",
				k,
				service,
				readOnlyFields,
			)
		}

		if !existsInCurConfig || curValue != v {
			updated = true
			curConfigToUpdate[k] = v
		}
	}

	return updated, http.StatusOK, nil
}

func (h *EditResourceHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, errors.Wrap(err, "Unable to edit resource.")
	}

	resourceIDStr := chi.URLParam(r, routes.IntegrationIdUrlParam)
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
		return nil, http.StatusBadRequest, errors.Wrap(err, "Error validating resource ownership.")
	}

	if !hasPermission {
		return nil, http.StatusForbidden, errors.New("You don't have permission to edit this resource")
	}

	name, configMap, err := request.ParseResourceConfigFromRequest(r)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to edit resource.")
	}

	if name == shared.DemoDbName {
		return nil, http.StatusBadRequest, errors.New("`aqueduct_demo` is reserved for demo resource. Please use another name.")
	}

	return &EditResourceArgs{
		AqContext:     aqContext,
		ResourceID:    resourceID,
		Name:          name,
		UpdatedFields: configMap,
	}, http.StatusOK, nil
}

func (h *EditResourceHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*EditResourceArgs)
	ID := args.ResourceID

	emptyResp := EditResourceResponse{}

	resourceObject, err := h.ResourceRepo.Get(ctx, ID, h.Database)
	if errors.Is(err, database.ErrNoRows()) {
		return emptyResp, http.StatusBadRequest, err
	}

	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to retrieve resource")
	}

	if resourceObject.Name == shared.DemoDbName {
		return emptyResp, http.StatusBadRequest, errors.New("You cannot edit demo DB credentials.")
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

	staticConfig, ok := config.(*auth.StaticConfig)
	if !ok {
		return emptyResp, http.StatusInternalServerError, errors.New("Editing for this resource type is not currently supported.")
	}

	configUpdated, status, err := updateConfig(staticConfig.Conf, resourceObject.Service, args.UpdatedFields)
	if err != nil {
		// Do not wrap err here since `updateConfig` returns a proper top-level message.
		return emptyResp, status, err
	}

	if !configUpdated {
		// handle name update if necessary:
		if args.Name != "" && args.Name != resourceObject.Name {
			status, err = UpdateResource(
				ctx,
				resourceObject.ID,
				args.Name,
				nil,
				h.ResourceRepo,
				h.Database,
				vaultObject,
			)
			if err != nil {
				return emptyResp, status, err
			}
		}

		return emptyResp, http.StatusOK, nil
	}

	// Validate resource config
	statusCode, err := ValidateConfig(
		ctx,
		args.RequestID,
		staticConfig,
		resourceObject.Service,
		h.JobManager,
		args.StorageConfig,
	)
	if err != nil {
		return emptyResp, statusCode, err
	}

	if statusCode, err := UpdateResource(
		ctx,
		resourceObject.ID,
		args.Name,
		staticConfig,
		h.ResourceRepo,
		h.Database,
		vaultObject,
	); err != nil {
		return emptyResp, statusCode, err
	}

	return emptyResp, http.StatusOK, nil
}

// UpdateResource updates an existing resource
// given the `newName` and / or `newConfig`.

func UpdateResource(
	ctx context.Context,
	resourceID uuid.UUID,
	newName string,
	newConfig auth.Config,
	resourceRepo repos.Resource,
	DB database.Database,
	vaultObject vault.Vault,
) (int, error) {
	changedFields := make(map[string]interface{}, 2)
	if newName != "" {
		changedFields[models.ResourceName] = newName
	}

	if newConfig != nil {
		// Extract non-confidential config
		publicConfig := newConfig.PublicConfig()
		changedFields[models.ResourceConfig] = (*shared.ResourceConfig)(&publicConfig)
	}

	txn, err := DB.BeginTx(ctx)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to update resource.")
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	_, err = resourceRepo.Update(
		ctx,
		resourceID,
		changedFields,
		txn,
	)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to update resource.")
	}

	// Store config (including confidential information) as in vault
	if newConfig != nil {
		if err := auth.WriteConfigToSecret(
			ctx,
			resourceID,
			newConfig,
			vaultObject,
		); err != nil {
			return http.StatusInternalServerError, errors.Wrap(err, "Unable to update resource.")
		}
	}

	if err := txn.Commit(ctx); err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to update resource.")
	}

	return http.StatusOK, nil
}
