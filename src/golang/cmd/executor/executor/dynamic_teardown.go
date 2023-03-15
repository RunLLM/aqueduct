package executor

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/aqueducthq/aqueduct/lib/dynamic"
	"github.com/aqueducthq/aqueduct/lib/models"
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

	if len(dynamicIntegrations) == 0 {
		log.Info("No dynamic integration detected, exiting...")
		return nil
	}

	var wg sync.WaitGroup

	for i := range dynamicIntegrations {
		wg.Add(1) // increment the WaitGroup counter
		// We use go routines to delete the clusters in parallel.
		go func(dynamicIntegration *models.Integration) {
			log.Infof("Checking dynamic integration %s, whose terraform directory is %s", dynamicIntegration.Name, dynamicIntegration.Config[shared.K8sTerraformPathKey])
			defer wg.Done() // decrement the WaitGroup counter when the goroutine completes

			if err := dynamic.ResyncClusterState(ctx, dynamicIntegration, ex.IntegrationRepo, ex.Database); err != nil {
				log.Error(errors.Wrap(err, "Failed to resync cluster state"))
				return
			}

			if dynamicIntegration.Config[shared.K8sStatusKey] == string(shared.K8sClusterActiveStatus) {
				lastUsedTimestampStr := dynamicIntegration.Config[shared.K8sLastUsedTimestampKey]
				lastUsedTimestamp, err := strconv.ParseInt(lastUsedTimestampStr, 10, 64)
				if err != nil {
					log.Error(errors.Wrap(err, "Unable to cast last used timestamp to int64"))
					return
				}

				keepaliveStr := dynamicIntegration.Config["keepalive"]
				keepalive, err := strconv.ParseInt(keepaliveStr, 10, 64)
				if err != nil {
					log.Error(errors.Wrap(err, "Unable to cast keepalive period to int64"))
					return
				}

				currTimestamp := time.Now().Unix()
				if (currTimestamp - lastUsedTimestamp) > keepalive {
					log.Info("Reached keepalive threshold, tearing down the cluster...")
					if err = dynamic.DeleteK8sCluster(
						ctx,
						// Perform pods status check because in case there are still pods running, we don't
						// want them to be killed by the teardown cron job.
						false,
						dynamicIntegration,
						ex.IntegrationRepo,
						ex.Database,
					); err != nil {
						log.Error(errors.Wrap(err, "Unable to delete dynamic k8s integration"))
						return
					}
				} else {
					log.Info("Have not reached keepalive threshold, not tearing down the cluster.")
				}
			}
		}(&dynamicIntegrations[i])
	}

	wg.Wait() // wait for all the goroutines to complete
	return nil
}
