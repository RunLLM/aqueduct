package utils

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// MigrateStorage moves all storage content from `oldConf` to `newConf`.
// This includes:
//   - artifact result content
//   - operator (function, check) code
//
// If the migration is successful, the above content is deleted from `oldConf`.
func MigrateStorage(
	ctx context.Context,
	oldConf *shared.StorageConfig,
	newConf *shared.StorageConfig,
	dagReader workflow_dag.Reader,
	dagWriter workflow_dag.Writer,
	artifactReader artifact.Reader,
	artifactResultReader artifact_result.Reader,
	operatorReader operator.Reader,
	db database.Database,
) error {
	// Wait until there are no more workflow runs in progress
	lock := NewExecutionLock()
	if err := lock.Lock(); err != nil {
		return err
	}
	defer func() {
		unlockErr := lock.Unlock()
		if unlockErr != nil {
			logrus.Errorf("Unexpected error when unlocking workflow execution lock: %v", unlockErr)
		}
	}()

	oldStore := storage.NewStorage(oldConf)
	newStore := storage.NewStorage(newConf)

	txn, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	dags, err := dagReader.GetWorkflowDags(ctx, nil, txn)
	if err != nil {
		return err
	}

	toDelete := []string{}

	for _, dag := range dags {
		if dag.EngineConfig.Type == shared.AirflowEngineType {
			// We cannot migrate content for Airflow workflows
			continue
		}

		// Migrate all of the artifact result content for this DAG
		artifacts, err := artifactReader.GetArtifactsByWorkflowDagId(ctx, dag.Id, txn)
		if err != nil {
			return err
		}

		for _, artifact := range artifacts {
			artifactResults, err := artifactResultReader.GetArtifactResultsByArtifactId(ctx, artifact.Id, txn)
			if err != nil {
				return err
			}

			// For each artifact result, move the content from `oldStore` to `newStore`
			for _, artifactResult := range artifactResults {
				val, err := oldStore.Get(ctx, artifactResult.ContentPath)
				if err != nil {
					return err
				}

				if err := newStore.Put(ctx, artifactResult.ContentPath, val); err != nil {
					return err
				}

				toDelete = append(toDelete, artifactResult.ContentPath)
			}
		}

		// Migrate all operator code for this DAG
		operators, err := operatorReader.GetOperatorsByWorkflowDagId(ctx, dag.Id, txn)
		if err != nil {
			return err
		}

		for _, operator := range operators {
			var operatorCodePath string
			switch {
			case operator.Spec.IsFunction():
				operatorCodePath = operator.Spec.Function().StoragePath
			case operator.Spec.IsCheck():
				operatorCodePath = operator.Spec.Check().Function.StoragePath
			case operator.Spec.IsMetric():
				operatorCodePath = operator.Spec.Metric().Function.StoragePath
			default:
				// There is no operator code to migrate for this operator
				continue
			}

			val, err := oldStore.Get(ctx, operatorCodePath)
			if err != nil {
				return err
			}

			if err := newStore.Put(ctx, operatorCodePath, val); err != nil {
				return err
			}

			toDelete = append(toDelete, operatorCodePath)
		}

		// Update the storage config for the DAG
		if _, err := dagWriter.UpdateWorkflowDag(
			ctx,
			dag.Id,
			map[string]interface{}{
				workflow_dag.StorageConfigColumn: newConf,
			},
			txn,
		); err != nil {
			return err
		}
	}

	if err := txn.Commit(ctx); err != nil {
		return err
	}

	// Delete keys from `oldStore` now that everything is fully migrated to `newStore`
	for _, key := range toDelete {
		if err := oldStore.Delete(ctx, key); err != nil {
			log.Errorf("Unexpected error when deleting %v after storage migration: %v", key, err)
		}
	}

	return nil
}
