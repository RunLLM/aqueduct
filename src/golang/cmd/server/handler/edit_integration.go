package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/request"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
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

var (
	ErrNoEditPermission            = errors.New("You don't have permission to edit this integration")
	ErrInvalidServiceType          = errors.New("Editing for this integration type is not currently supported.")
	ErrEditDemoIntegration         = errors.New("You cannot edit demo DB credentials.")
	ErrEditIntegrationWithDemoName = errors.New("aqueduct_demo is reserved for demo integration. Please use another name.")
)

// ConnectIntegrationHandler connects a new integration for the organization.
type EditIntegrationHandler struct {
	PostHandler

	Database          database.Database
	IntegrationReader integration.Reader
	IntegrationWriter integration.Writer
	Vault             vault.Vault
	JobManager        job.JobManager
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
	IntegrationId uuid.UUID
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

	integrationIdStr := chi.URLParam(r, routes.IntegrationIdUrlParam)
	integrationId, err := uuid.Parse(integrationIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed integration ID.")
	}

	hasPermission, err := h.IntegrationReader.ValidateIntegrationOwnership(
		r.Context(),
		integrationId,
		aqContext.OrganizationId,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Error validating integration ownership.")
	}

	if !hasPermission {
		return nil, http.StatusForbidden, ErrNoEditPermission
	}

	name, configMap, err := request.ParseIntegrationConfigFromRequest(r)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to edit integration.")
	}

	if name == integration.DemoDbIntegrationName {
		return nil, http.StatusBadRequest, ErrEditIntegrationWithDemoName
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

	if integrationObject.Name == integration.DemoDbIntegrationName {
		return emptyResp, http.StatusBadRequest, ErrEditDemoIntegration
	}

	config, err := auth.ReadConfigFromSecret(ctx, id, h.Vault)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Failed to retrieve secrets")
	}

	staticConfig, ok := config.(*auth.StaticConfig)
	if !ok {
		return emptyResp, http.StatusInternalServerError, ErrInvalidServiceType
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
				integrationObject.Id,
				args.Name,
				nil,
				h.IntegrationWriter,
				h.Database,
				h.Vault,
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
		args.RequestId,
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
		integrationObject.Id,
		args.Name,
		staticConfig,
		h.IntegrationWriter,
		h.Database,
		h.Vault,
	); err != nil {
		return emptyResp, statusCode, err
	}

	return emptyResp, http.StatusOK, nil
}

// UpdateIntegration updates an existing integration
// given the `newName` and / or `newConfig`.

func UpdateIntegration(
	ctx context.Context,
	integrationId uuid.UUID,
	newName string,
	newConfig auth.Config,
	integrationWriter integration.Writer,
	db database.Database,
	vaultObject vault.Vault,
) (int, error) {
	changedFields := make(map[string]interface{}, 2)
	if newName != "" {
		changedFields[integration.NameColumn] = newName
	}

	if newConfig != nil {
		// Extract non-confidential config
		publicConfig := newConfig.PublicConfig()
		changedFields[integration.ConfigColumn] = (*postgres_utils.Config)(&publicConfig)
	}

	txn, err := db.BeginTx(ctx)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "Unable to update integration.")
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	_, err = integrationWriter.UpdateIntegration(
		ctx,
		integrationId,
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
			integrationId,
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
