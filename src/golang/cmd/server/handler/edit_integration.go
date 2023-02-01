package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/request"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Route: /integration/{integrationId}/edit
// Method: POST
// Params: integrationId
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
type EditIntegrationHandler struct {
	PostHandler

	Database   database.Database
	JobManager job.JobManager

	IntegrationRepo repos.Integration
}

var serviceToReadOnlyFields = map[integration.Service]map[string]bool{
	integration.Airflow:  {"host": true},
	integration.BigQuery: {"project_id": true},
	integration.MariaDb: {
		"host":     true,
		"port":     true,
		"database": true,
	},
	integration.MySql: {
		"host":     true,
		"port":     true,
		"database": true,
	},
	integration.Postgres: {
		"host":     true,
		"port":     true,
		"database": true,
	},
	integration.Redshift: {
		"host":     true,
		"port":     true,
		"database": true,
	},
	integration.S3: {
		"bucket":         true,
		"region":         true,
		"use_as_storage": true,
	},
	integration.Snowflake: {
		"account_identifier": true,
		"warehouse":          true,
		"database":           true,
	},
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
	IntegrationID uuid.UUID
	UpdatedFields map[string]string
}

type EditIntegrationResponse struct{}

func (*EditIntegrationHandler) Name() string {
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
	service integration.Service,
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

func (h *EditIntegrationHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, errors.Wrap(err, "Unable to edit integration.")
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
		return nil, http.StatusBadRequest, errors.Wrap(err, "Error validating integration ownership.")
	}

	if !hasPermission {
		return nil, http.StatusForbidden, errors.New("You don't have permission to edit this integration")
	}

	name, configMap, err := request.ParseIntegrationConfigFromRequest(r)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to edit integration.")
	}

	if name == integration.DemoDbIntegrationName {
		return nil, http.StatusBadRequest, errors.New("`aqueduct_demo` is reserved for demo integration. Please use another name.")
	}

	return &EditIntegrationArgs{
		AqContext:     aqContext,
		IntegrationID: integrationID,
		Name:          name,
		UpdatedFields: configMap,
	}, http.StatusOK, nil
}

func (h *EditIntegrationHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*EditIntegrationArgs)
	ID := args.IntegrationID

	emptyResp := EditIntegrationResponse{}

	integrationObject, err := h.IntegrationRepo.Get(ctx, ID, h.Database)
	if err == database.ErrNoRows {
		return emptyResp, http.StatusBadRequest, err
	}

	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to retrieve integration")
	}

	if integrationObject.Name == integration.DemoDbIntegrationName {
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
		return emptyResp, http.StatusInternalServerError, errors.New("Editing for this integration type is not currently supported.")
	}

	configUpdated, status, err := updateConfig(staticConfig.Conf, integrationObject.Service, args.UpdatedFields)
	if err != nil {
		// Do not wrap err here since `updateConfig` returns a proper top-level message.
		return emptyResp, status, err
	}

	if !configUpdated {
		// handle name update if necessary:
		if args.Name != "" && args.Name != integrationObject.Name {
			status, err = UpdateIntegration(
				ctx,
				integrationObject.ID,
				args.Name,
				nil,
				h.IntegrationRepo,
				h.Database,
				vaultObject,
			)
			if err != nil {
				return emptyResp, status, err
			}
		}

		return emptyResp, http.StatusOK, nil
	}

	// Validate integration config
	statusCode, err := ValidateConfig(
		ctx,
		args.RequestID,
		staticConfig,
		integrationObject.Service,
		h.JobManager,
		args.StorageConfig,
	)
	if err != nil {
		return emptyResp, statusCode, err
	}

	if statusCode, err := UpdateIntegration(
		ctx,
		integrationObject.ID,
		args.Name,
		staticConfig,
		h.IntegrationRepo,
		h.Database,
		vaultObject,
	); err != nil {
		return emptyResp, statusCode, err
	}

	return emptyResp, http.StatusOK, nil
}

// UpdateIntegration updates an existing integration
// given the `newName` and / or `newConfig`.

func UpdateIntegration(
	ctx context.Context,
	integrationID uuid.UUID,
	newName string,
	newConfig auth.Config,
	integrationRepo repos.Integration,
	DB database.Database,
	vaultObject vault.Vault,
) (int, error) {
	changedFields := make(map[string]interface{}, 2)
	if newName != "" {
		changedFields[models.IntegrationName] = newName
	}

	if newConfig != nil {
		// Extract non-confidential config
		publicConfig := newConfig.PublicConfig()
		changedFields[models.IntegrationConfig] = (*utils.Config)(&publicConfig)
	}

	txn, err := DB.BeginTx(ctx)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to update integration.")
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	_, err = integrationRepo.Update(
		ctx,
		integrationID,
		changedFields,
		txn,
	)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to update integration.")
	}

	// Store config (including confidential information) as in vault
	if newConfig != nil {
		if err := auth.WriteConfigToSecret(
			ctx,
			integrationID,
			newConfig,
			vaultObject,
		); err != nil {
			return http.StatusInternalServerError, errors.Wrap(err, "Unable to update integration.")
		}
	}

	if err := txn.Commit(ctx); err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to update integration.")
	}

	return http.StatusOK, nil
}
