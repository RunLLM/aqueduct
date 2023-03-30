package storage_migration

import (
	"context"

	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/sirupsen/logrus"
)

// StorageCleanupConfig contains the fields necessary to cleanup the old storage layer.
// Callers of `MigrateStorageAndVault` can use this config struct to perform best-effort cleanup
// after the migration completes.
type StorageCleanupConfig struct {
	StoreKeys []string
	VaultKeys []string

	Store storage.Storage
	Vault vault.Vault
}

// MigrateStorageAndVault copies all storage (and vault) content from `oldConf` to `newConf`.
// This includes:
//   - artifact result content
//   - operator (function, check) code
//   - vault content (integration credentials)
//
// The keys to all the contents that were copied are also returned, so that the caller can perform best-effort
// cleanup the old storage layer.
func MigrateStorageAndVault(
	ctx context.Context,
	oldConf *shared.StorageConfig,
	newConf *shared.StorageConfig,
	orgID string,
	dagRepo repos.DAG,
	artifactRepo repos.Artifact,
	artifactResultRepo repos.ArtifactResult,
	operatorRepo repos.Operator,
	integrationRepo repos.Integration,
	DB database.Database,
) (*StorageCleanupConfig, error) {
	logrus.Infof("Migrating from %v to %v", *oldConf, *newConf)

	oldStore := storage.NewStorage(oldConf)
	newStore := storage.NewStorage(newConf)

	oldVault, err := vault.NewVault(oldConf, config.EncryptionKey())
	if err != nil {
		return nil, err
	}

	newVault, err := vault.NewVault(newConf, config.EncryptionKey())
	if err != nil {
		return nil, err
	}

	txn, err := DB.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	dags, err := dagRepo.List(ctx, txn)
	if err != nil {
		return nil, err
	}

	toDelete := []string{}

	logrus.Infof("There are %v DAGs to migrate", len(dags))

	for _, dag := range dags {
		logrus.Infof("Starting migration for DAG %v", dag.ID)

		if dag.EngineConfig.Type == shared.AirflowEngineType {
			// We cannot migrate content for Airflow workflows
			logrus.Info("This DAG's engine is Airflow, so its migration will be skipped.")
			continue
		}

		// Migrate all of the artifact result content for this DAG
		artifacts, err := artifactRepo.GetByDAG(ctx, dag.ID, txn)
		if err != nil {
			return nil, err
		}

		logrus.Infof("There are %v artifacts to migrate for DAG %v", len(artifacts), dag.ID)

		for _, artifact := range artifacts {
			logrus.Infof("Starting migration for artifact %v of DAG %v", artifact.ID, dag.ID)

			artifactResults, err := artifactResultRepo.GetByArtifact(ctx, artifact.ID, txn)
			if err != nil {
				return nil, err
			}

			logrus.Infof("There are %v artifact results to migrate for artifact %v", len(artifactResults), artifact.ID)

			// For each artifact result, move the content from `oldStore` to `newStore`
			for _, artifactResult := range artifactResults {
				logrus.Infof("Starting migration for artifact result %v of artifact %v", artifactResult.ID, artifact.ID)

				val, err := oldStore.Get(ctx, artifactResult.ContentPath)
				if err != nil &&
					!artifactResult.ExecState.IsNull &&
					artifactResult.ExecState.Status == shared.SucceededExecutionStatus {
					// Return an error because the artifact result is successful,
					// but not found in current storage.
					logrus.Errorf("Unable to get artifact result %v from old store: %v", artifactResult.ID, err)
					return nil, err
				}

				if err == nil {
					// Only try to migrate artifact result if there was no issue reading
					// it from the `oldStore`
					if err := newStore.Put(ctx, artifactResult.ContentPath, val); err != nil {
						logrus.Errorf("Unable to write artifact result %v to new store: %v", artifactResult.ID, err)
						return nil, err
					}
				}

				toDelete = append(toDelete, artifactResult.ContentPath)
			}
		}

		// Migrate all operator code for this DAG
		operators, err := operatorRepo.GetByDAG(ctx, dag.ID, txn)
		if err != nil {
			return nil, err
		}

		logrus.Infof("There are %v operators to migrate for DAG %v", len(operators), dag.ID)

		for _, operator := range operators {
			logrus.Infof("Starting migration for operator %v of DAG %v", operator.ID, dag.ID)

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
				logrus.Errorf("Unable to get operator code %v from old store: %v", operator.ID, err)
				return nil, err
			}

			if err := newStore.Put(ctx, operatorCodePath, val); err != nil {
				logrus.Errorf("Unable to write operator code %v to new store: %v", operator.ID, err)
				return nil, err
			}

			toDelete = append(toDelete, operatorCodePath)
		}

		// Update the storage config for the DAG
		if _, err := dagRepo.Update(
			ctx,
			dag.ID,
			map[string]interface{}{
				models.DagStorageConfig: newConf,
			},
			txn,
		); err != nil {
			return nil, err
		}
	}

	// Migrate the vault portion of storage
	toDeleteFromVault, err := utils.MigrateVault(
		ctx,
		oldVault,
		newVault,
		orgID,
		integrationRepo,
		txn,
	)
	if err != nil {
		return nil, err
	}

	if err := txn.Commit(ctx); err != nil {
		return nil, err
	}

	return &StorageCleanupConfig{
		StoreKeys: toDelete,
		VaultKeys: toDeleteFromVault,
		Store:     oldStore,
		Vault:     oldVault,
	}, nil
}
