package executor

import (
	"context"
	"strconv"
	"time"

	"github.com/aqueducthq/aqueduct/lib/dynamic"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
)

type DynamicTeardownExecutor struct {
	*BaseExecutor
}

func NewDynamicTeardownExecutor(base *BaseExecutor) *DynamicTeardownExecutor {
	return &DynamicTeardownExecutor{BaseExecutor: base}
}

// Run inspects each dynamic integration and tears down the cluster if it has been idle for a while.
// This check is performed by subtracting the last-updated-timestamp from the current timestamp and
// comparing it with the keepalive threshold. The last-used-timestamp is updated whenever an operator
// makes uses of the dynamic integration.
func (ex *DynamicTeardownExecutor) Run(ctx context.Context) error {
	log.Info("Starting dynamic integration teardown.")

	dynamicIntegrations, err := ex.IntegrationRepo.GetByConfigField(ctx, shared.K8sDynamicKey, strconv.FormatBool(true), ex.Database)
	if err != nil {
		return errors.Wrap(err, "Unable to get dynamic integration.")
	}

	if len(dynamicIntegrations) > 1 {
		return errors.New("Got more than one dynamic integration. Currently this should never happen.")
	}

	if len(dynamicIntegrations) == 0 {
		log.Info("No dynamic integration detected, exiting...")
		return nil
	}

	dynamicIntegration := dynamicIntegrations[0]

	if err := dynamic.ResyncClusterState(ctx, &dynamicIntegration, ex.IntegrationRepo, ex.Database); err != nil {
		return errors.Wrap(err, "Failed to resync cluster state")
	}

	if dynamicIntegration.Config[shared.K8sStatusKey] == string(shared.K8sClusterActiveStatus) {
		lastUsedTimestampStr := dynamicIntegration.Config[shared.K8sLastUsedTimestampKey]
		lastUsedTimestamp, err := strconv.ParseInt(lastUsedTimestampStr, 10, 64)
		if err != nil {
			return errors.Wrap(err, "Unable to cast last used timestamp to int64")
		}

		keepaliveStr := dynamicIntegration.Config["keepalive"]
		keepalive, err := strconv.ParseInt(keepaliveStr, 10, 64)
		if err != nil {
			return errors.Wrap(err, "Unable to cast keepalive period to int64")
		}

		currTimestamp := time.Now().Unix()
		if (currTimestamp - lastUsedTimestamp) > keepalive {
			log.Info("Reached keepalive threshold, tearing down the cluster...")
			if err = dynamic.DeleteK8sCluster(
				ctx,
				false, // don't force deletion if there're pods still running
				&dynamicIntegration,
				ex.IntegrationRepo,
				ex.Database,
			); err != nil {
				return errors.Wrap(err, "Unable to delete dynamic k8s integration")
			}
		} else {
			log.Info("Have not reached keepalive threshold, not tearing down the cluster.")
		}
	}

	return nil
}
