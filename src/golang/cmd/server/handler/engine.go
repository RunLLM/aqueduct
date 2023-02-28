package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/aqueducthq/aqueduct/cmd/server/response"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/config"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// Route: /api/integration/engine/{integrationId}
// Method: POST/DELETE
// Params:
//
//	`integrationId`: ID of the dynamic engine integration
//
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//
// Response: serialized `GetEngineStatusResponse`.
type EngineHandler struct {
	PostHandler

	Database database.Database

	IntegrationRepo repos.Integration
}

type engineArgs struct {
	*aq_context.AqContext
	action        string
	integrationId uuid.UUID
}

func (*EngineHandler) Name() string {
	return "Engine"
}

func (*EngineHandler) Headers() []string {
	return []string{
		routes.EngineActionHeader,
	}
}

func (*EngineHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	integrationIdStr := chi.URLParam(r, routes.IntegrationIdUrlParam)
	integrationId, err := uuid.Parse(integrationIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed engine integration ID.")
	}

	return &engineArgs{
		AqContext:     aqContext,
		action:        r.Header.Get("action"),
		integrationId: integrationId,
	}, http.StatusOK, nil
}

func (h *EngineHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*engineArgs)
	log.Infof("!!!!!!!Method is %s!!!!!!", args.action)

	emptyResponse := response.EmptyResponse{}

	engineIntegration, err := h.IntegrationRepo.Get(
		ctx,
		args.integrationId,
		h.Database,
	)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to get engine integration.")
	}

	if _, ok := engineIntegration.Config["dynamic"]; !ok {
		return emptyResponse, http.StatusBadRequest, errors.New("This is not a dynamic engine integration.")
	}

	if args.action == "create" {
		// This is a cluster creation request.
		engineIntegration, err := engine.UpdateEngineLastUsedTimestamp(
			ctx,
			args.integrationId,
			h.IntegrationRepo,
			h.Database,
		)
		if err != nil {
			return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Failed to update engine last used timestamp")
		}

		storageConfig := config.Storage()
		vaultObject, err := vault.NewVault(&storageConfig, config.EncryptionKey())
		if err != nil {
			return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to initialize vault.")
		}

		for {
			if engineIntegration.Config["status"] == string(shared.K8sClusterTerminatedStatus) {
				log.Info("engine is currently terminated, starting...")
				err = engine.CreateDynamicEngine(
					ctx,
					engineIntegration,
					h.IntegrationRepo,
					vaultObject,
					h.Database,
				)
				if err != nil {
					return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to create dynamic engine")
				}
				return emptyResponse, http.StatusOK, nil
			} else if engineIntegration.Config["status"] == string(shared.K8sClusterActiveStatus) {
				return emptyResponse, http.StatusOK, nil
			} else {
				log.Infof("Kubernetes cluster is currently in %s status. Waiting for 30 seconds before checking again...", engineIntegration.Config["status"])
				time.Sleep(30 * time.Second)

				engineIntegration, err = h.IntegrationRepo.Get(
					ctx,
					args.integrationId,
					h.Database,
				)
				if err != nil {
					return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Failed to retrieve engine integration")
				}
			}
		}
	} else if args.action == "delete" {
		// This is a cluster deletion request.
		for {
			if engineIntegration.Config["status"] == string(shared.K8sClusterActiveStatus) {
				log.Info("Tearing down the cluster...")
				if err = engine.DeleteDynamicEngine(
					ctx,
					engineIntegration,
					h.IntegrationRepo,
					h.Database,
				); err != nil {
					return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to delete dynamic k8s integration")
				}

				return emptyResponse, http.StatusOK, nil
			} else if engineIntegration.Config["status"] == string(shared.K8sClusterTerminatedStatus) {
				return emptyResponse, http.StatusOK, nil
			} else {
				log.Infof("Kubernetes cluster is currently in %s status. Waiting for 30 seconds before checking again...", engineIntegration.Config["status"])
				time.Sleep(30 * time.Second)

				engineIntegration, err = h.IntegrationRepo.Get(
					ctx,
					args.integrationId,
					h.Database,
				)
				if err != nil {
					return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Failed to retrieve engine integration")
				}
			}
		}
	} else {
		return emptyResponse, http.StatusBadRequest, errors.Newf("Unsupported action: %s.", args.action)
	}
}
