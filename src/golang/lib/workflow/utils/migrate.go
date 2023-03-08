package utils

import (
	"context"

	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/aqueducthq/aqueduct/lib/vault"
	log "github.com/sirupsen/logrus"
)

// MigrateStorageAndVault moves all storage (and vault) content from `oldConf` to `newConf`.
// This includes:
//   - artifact result content
//   - operator (function, check) code
//   - vault content (integration credentials)
//
// If the migration is successful, the above content is deleted from `oldConf`.
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
) error {
	log.Infof("Migrating from %v to %v", *oldConf, *newConf)

	oldStore := storage.NewStorage(oldConf)
	newStore := storage.NewStorage(newConf)

	oldVault, err := vault.NewVault(oldConf, config.EncryptionKey())
	if err != nil {
		return err
	}

	newVault, err := vault.NewVault(newConf, config.EncryptionKey())
	if err != nil {
		return err
	}

	txn, err := DB.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	dags, err := dagRepo.List(ctx, txn)
	if err != nil {
		return err
	}

	toDelete := []string{}

	log.Infof("There are %v DAGs to migrate", len(dags))

	for _, dag := range dags {
		log.Infof("Starting migration for DAG %v", dag.ID)

		if dag.EngineConfig.Type == shared.AirflowEngineType {
			// We cannot migrate content for Airflow workflows
			log.Info("This DAG's engine is Airflow, so its migration will be skipped.")
			continue
		}

		// Migrate all of the artifact result content for this DAG
		artifacts, err := artifactRepo.GetByDAG(ctx, dag.ID, txn)
		if err != nil {
			return err
		}

		log.Infof("There are %v artifacts to migrate for DAG %v", len(artifacts), dag.ID)

		for _, artifact := range artifacts {
			log.Infof("Starting migration for artifact %v of DAG %v", artifact.ID, dag.ID)

			artifactResults, err := artifactResultRepo.GetByArtifact(ctx, artifact.ID, txn)
			if err != nil {
				return err
			}

			log.Infof("There are %v artifact results to migrate for artifact %v", len(artifactResults), artifact.ID)

			// For each artifact result, move the content from `oldStore` to `newStore`
			for _, artifactResult := range artifactResults {
				log.Infof("Starting migration for artifact result %v of artifact %v", artifactResult.ID, artifact.ID)

				val, err := oldStore.Get(ctx, artifactResult.ContentPath)
				if err != nil &&
					!artifactResult.ExecState.IsNull &&
					artifactResult.ExecState.Status == shared.SucceededExecutionStatus {
					// Return an error because the artifact result is successful,
					// but not found in current storage.
					log.Errorf("Unable to get artifact result %v from old store: %v", artifactResult.ID, err)
					return err
				}

				if err == nil {
					// Only try to migrate artifact result if there was no issue reading
					// it from the `oldStore`
					if err := newStore.Put(ctx, artifactResult.ContentPath, val); err != nil {
						log.Errorf("Unable to write artifact result %v to new store: %v", artifactResult.ID, err)
						return err
					}
				}

				toDelete = append(toDelete, artifactResult.ContentPath)
			}
		}

		// Migrate all operator code for this DAG
		operators, err := operatorRepo.GetByDAG(ctx, dag.ID, txn)
		if err != nil {
			return err
		}

		log.Infof("There are %v operators to migrate for DAG %v", len(operators), dag.ID)

		for _, operator := range operators {
			log.Infof("Starting migration for operator %v of DAG %v", operator.ID, dag.ID)

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
				log.Errorf("Unable to get operator code %v from old store: %v", operator.ID, err)
				return err
			}

			if err := newStore.Put(ctx, operatorCodePath, val); err != nil {
				log.Errorf("Unable to write operator code %v to new store: %v", operator.ID, err)
				return err
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
			return err
		}
	}

	// Migrate the vault portion of storage
	toDeleteFromVault, err := MigrateVault(
		ctx,
		oldVault,
		newVault,
		orgID,
		integrationRepo,
		txn,
	)
	if err != nil {
		return err
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

	// Delete keys from `oldVault` now that everything is fully migrated to `newVault`
	for _, key := range toDeleteFromVault {
		if err := oldVault.Delete(ctx, key); err != nil {
			log.Errorf("Unexpected error when deleting %v after vault migration: %v", key, err)
		}
	}

	return nil
}

// MigrateVault migrates all vault content from `oldVault` to `newVault`.
// This includes:
//   - integration credentials
//
// It also returns the names of all the keys that have been migrated to `newVault`.
// It is the responsibility of the caller to delete the keys if necessary.
func MigrateVault(
	ctx context.Context,
	oldVault vault.Vault,
	newVault vault.Vault,
	orgID string,
	integrationRepo repos.Integration,
	DB database.Database,
) ([]string, error) {
	integrations, err := integrationRepo.GetByOrg(ctx, orgID, DB)
	if err != nil {
		return nil, err
	}

	keys := []string{}

	log.Infof("There are %v integrations to migrate", len(integrations))

	// For each connected integration, migrate its credentials
	for _, integrationDB := range integrations {
		log.Infof("Starting migration for integration %v %v", integrationDB.ID, integrationDB.Name)
		// The vault key for the credentials is the integration record's ID
		key := integrationDB.ID.String()

		val, err := oldVault.Get(ctx, key)
		if err != nil {
			log.Errorf("Unable to get integration credentials %v from old vault: %v", integrationDB.ID, err)
			return nil, err
		}

		if err := newVault.Put(ctx, key, val); err != nil {
			log.Errorf("Unable to write integration credentials %v to new vault: %v", integrationDB.ID, err)
			return nil, err
		}

		keys = append(keys, key)
	}

	return keys, nil
}
