package artifact

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/sys/filelock"
	"github.com/sirupsen/logrus"
)

// Migrate moves from workflow artifact result content from `oldConf` to `newConf`.
func Migrate(
	ctx context.Context,
	oldConf *shared.StorageConfig,
	newConf *shared.StorageConfig,
	dagReader workflow_dag.Reader,
	dagWriter workflow_dag.Writer,
	artifactReader artifact.Reader,
	artifactResultReader artifact_result.Reader,
	db database.Database,
) error {
	// Wait until there are no more workflow runs in progress
	lock := filelock.New(utils.ExecutionLock)
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

	for _, dag := range dags {
		if dag.EngineConfig.Type == shared.AirflowEngineType {
			// We cannot migrate content for Airflow workflows
			continue
		}

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
			}
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

	return nil
}
