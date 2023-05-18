package handler

import (
	"context"
	"encoding/json"
	"net/http"

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

// Route: /api/integration/dynamic-engine/{resourceID}/edit
// Method: POST
// Params:
//
//	`resourceID`: ID of the dynamic engine integration
//
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//		`action`: indicates whether this is a creation or deletion request
type EditDynamicEngineHandler struct {
	PostHandler

	Database database.Database

	ResourceRepo repos.Resource
}

type editDynamicEngineArgs struct {
	*aq_context.AqContext
	action      dynamicEngineAction
	resourceID  uuid.UUID
	configDelta *shared.DynamicK8sConfig
}

func (*EditDynamicEngineHandler) Name() string {
	return "EditDynamicEngine"
}

func (*EditDynamicEngineHandler) Headers() []string {
	return []string{
		routes.DynamicEngineActionHeader,
	}
}

type dynamicEngineAction string

const (
	// These reflect K8sClusterActionType in Python and should be kept in sync.
	createAction      dynamicEngineAction = "create"
	updateAction      dynamicEngineAction = "update"
	deleteAction      dynamicEngineAction = "delete"
	forceDeleteAction dynamicEngineAction = "force-delete"
	// The config delta payload sent from the client is keyed under this key in the HTTP request body.
	configDeltaKey string = "config_delta"
)

func isValidAction(action string) bool {
	switch dynamicEngineAction(action) {
	case createAction, updateAction, deleteAction, forceDeleteAction:
		return true
	default:
		return false
	}
}

func (*EditDynamicEngineHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	integrationIdStr := chi.URLParam(r, routes.ResourceIDUrlParam)
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
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to extract config delta.")
	}

	configDelta := shared.DynamicK8sConfig{}
	if err = json.Unmarshal(configDeltaBytes, &configDelta); err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to deserialize config delta.")
	}

	if !isValidAction(action) {
		return nil, http.StatusBadRequest, errors.Newf("Unsupported action: %s.", action)
	}

	return &editDynamicEngineArgs{
		AqContext:   aqContext,
		action:      dynamicEngineAction(action),
		resourceID:  integrationId,
		configDelta: &configDelta,
	}, http.StatusOK, nil
}

func (h *EditDynamicEngineHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*editDynamicEngineArgs)
	emptyResponse := response.EmptyResponse{}

	dynamicEngineIntegration, err := h.ResourceRepo.Get(
		ctx,
		args.resourceID,
		h.Database,
	)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to get dynamic engine integration.")
	}

	if _, ok := dynamicEngineIntegration.Config[shared.K8sDynamicKey]; !ok {
		return emptyResponse, http.StatusBadRequest, errors.New("This is not a dynamic engine integration.")
	}

	storageConfig := config.Storage()
	vaultObject, err := vault.NewVault(&storageConfig, config.EncryptionKey())
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to initialize vault.")
	}

	if args.action == createAction {
		log.Info("Received a cluster creation request")
		err = dynamic.PrepareCluster(
			ctx,
			args.configDelta,
			args.resourceID,
			h.ResourceRepo,
			vaultObject,
			h.Database,
		)
		if err != nil {
			return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to create dynamic k8s engine.")
		}

		return emptyResponse, http.StatusOK, nil
	} else if args.action == updateAction {
		log.Info("Received a cluster update request")
		if len(args.configDelta.ToMap()) == 0 {
			return emptyResponse, http.StatusBadRequest, errors.New("Empty config delta provided.")
		}

		if dynamicEngineIntegration.Config[shared.K8sStatusKey] != string(shared.K8sClusterActiveStatus) {
			return emptyResponse, http.StatusUnprocessableEntity, errors.Newf(
				"Action %s is only applicable when the cluster is in %s status, but it is now in %s status.",
				updateAction,
				shared.K8sClusterActiveStatus,
				dynamicEngineIntegration.Config[shared.K8sStatusKey],
			)
		}

		if err = dynamic.CreateOrUpdateK8sCluster(
			ctx,
			args.configDelta,
			dynamic.K8sClusterUpdateAction,
			dynamicEngineIntegration,
			h.ResourceRepo,
			vaultObject,
			h.Database,
		); err != nil {
			return emptyResponse, http.StatusInternalServerError, err
		}

		return emptyResponse, http.StatusOK, nil
	} else if args.action == deleteAction || args.action == forceDeleteAction {
		log.Info("Received a cluster deletion request")
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
					h.ResourceRepo,
					vaultObject,
					h.Database,
				); err != nil {
					return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to delete dynamic k8s engine")
				}

				return emptyResponse, http.StatusOK, nil
			} else if dynamicEngineIntegration.Config[shared.K8sStatusKey] == string(shared.K8sClusterTerminatedStatus) {
				return emptyResponse, http.StatusOK, nil
			} else {
				dynamicEngineIntegration, err = dynamic.PollClusterStatus(ctx, dynamicEngineIntegration, h.ResourceRepo, vaultObject, h.Database)
				if err != nil {
					return emptyResponse, http.StatusInternalServerError, err
				}
			}
		}
	} else {
		return emptyResponse, http.StatusBadRequest, errors.Newf("Unsupported action: %s.", args.action)
	}
}
