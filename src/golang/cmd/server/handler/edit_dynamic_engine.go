package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aqueducthq/aqueduct/cmd/server/request"
	"github.com/aqueducthq/aqueduct/cmd/server/response"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/config"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/dynamic"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// Route: /api/integration/dynamic-engine/{integrationId}/edit
// Method: POST
// Params:
//
//	`integrationId`: ID of the dynamic engine integration
//
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//		`action`: indicates whether this is a creation or deletion request
type EditDynamicEngineHandler struct {
	PostHandler

	Database database.Database

	IntegrationRepo repos.Integration
}

type editDynamicEngineArgs struct {
	*aq_context.AqContext
	action        string
	integrationId uuid.UUID
	configDelta   map[string]string
}

func (*EditDynamicEngineHandler) Name() string {
	return "EditDynamicEngine"
}

func (*EditDynamicEngineHandler) Headers() []string {
	return []string{
		routes.DynamicEngineActionHeader,
	}
}

const (
	createAction      string = "create"
	deleteAction      string = "delete"
	forceDeleteAction string = "force-delete"
	configDeltaKey    string = "config_delta"
)

func (*EditDynamicEngineHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	integrationIdStr := chi.URLParam(r, routes.IntegrationIdUrlParam)
	integrationId, err := uuid.Parse(integrationIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed dynamic engine integration ID.")
	}

	action := r.Header.Get("action")
	if action == "" {
		return nil, http.StatusBadRequest, errors.Wrap(err, "No action specified by the request.")
	}

	configDeltaBytes, err := request.ExtractHttpPayload(
		r.Header.Get(routes.ContentTypeHeader),
		configDeltaKey,
		false, // not a file
		r,
	)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to extract config delta map.")
	}

	configDelta := make(map[string]string)
	if err = json.Unmarshal(configDeltaBytes, &configDelta); err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to deserialize config delta map.")
	}

	return &editDynamicEngineArgs{
		AqContext:     aqContext,
		action:        action,
		integrationId: integrationId,
		configDelta:   configDelta,
	}, http.StatusOK, nil
}

func (h *EditDynamicEngineHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*editDynamicEngineArgs)
	emptyResponse := response.EmptyResponse{}

	dynamicEngineIntegration, err := h.IntegrationRepo.Get(
		ctx,
		args.integrationId,
		h.Database,
	)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to get dynamic engine integration.")
	}

	if _, ok := dynamicEngineIntegration.Config[shared.K8sDynamicKey]; !ok {
		return emptyResponse, http.StatusBadRequest, errors.New("This is not a dynamic engine integration.")
	}

	if args.action == createAction {
		// This is a cluster creation request.
		storageConfig := config.Storage()
		vaultObject, err := vault.NewVault(&storageConfig, config.EncryptionKey())
		if err != nil {
			return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to initialize vault.")
		}

		err = dynamic.PrepareEngine(
			ctx,
			args.configDelta,
			args.integrationId,
			h.IntegrationRepo,
			vaultObject,
			h.Database,
		)
		if err != nil {
			return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to create dynamic k8s engine.")
		}

		return emptyResponse, http.StatusOK, nil
	} else if args.action == deleteAction || args.action == forceDeleteAction {
		// This is a cluster deletion request.
		forceDelete := false
		if args.action == forceDeleteAction {
			forceDelete = true
		}

		for {
			if dynamicEngineIntegration.Config[shared.K8sStatusKey] == string(shared.K8sClusterActiveStatus) {
				log.Info("Tearing down the Kubernetes cluster...")
				if err = dynamic.DeleteK8sCluster(
					ctx,
					forceDelete,
					dynamicEngineIntegration,
					h.IntegrationRepo,
					h.Database,
				); err != nil {
					return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to delete dynamic k8s engine")
				}

				return emptyResponse, http.StatusOK, nil
			} else if dynamicEngineIntegration.Config[shared.K8sStatusKey] == string(shared.K8sClusterTerminatedStatus) {
				return emptyResponse, http.StatusOK, nil
			} else {
				if err := dynamic.ResyncClusterState(ctx, dynamicEngineIntegration, h.IntegrationRepo, h.Database); err != nil {
					return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Failed to resync cluster state")
				}

				if dynamicEngineIntegration.Config[shared.K8sStatusKey] == string(shared.K8sClusterTerminatedStatus) {
					// This means the cluster state is resynced to Terminated, so no need to wait 30 seconds.
					continue
				}

				log.Infof("Kubernetes cluster is currently in %s status. Waiting for 30 seconds before checking again...", dynamicEngineIntegration.Config[shared.K8sStatusKey])
				time.Sleep(30 * time.Second)

				dynamicEngineIntegration, err = h.IntegrationRepo.Get(
					ctx,
					args.integrationId,
					h.Database,
				)
				if err != nil {
					return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Failed to retrieve dynamic engine integration")
				}
			}
		}
	} else {
		return emptyResponse, http.StatusBadRequest, errors.Newf("Unsupported action: %s.", args.action)
	}
}
