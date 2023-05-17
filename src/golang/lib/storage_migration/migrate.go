package storage_migration

import (
	"context"
	"fmt"
	"time"

	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/database"
	aq_errors "github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// Perform starts and manages the storage migration process through its entire lifecycle.
// The migration logic is performed asynchronously, and it's process is tracked as a new entry in the
// `storage_migration` table. This method does not block on the migration process.
// This method also pauses the server, until the migration completes or errors, wherein the server will restart.
func Perform(
	ctx context.Context,
	orgID string,
	destResourceObj *models.Resource,
	newStorageConfig *shared.StorageConfig,
	pauseServer func(),
	restartServer func(),
	artifactRepo repos.Artifact,
	artifactResultRepo repos.ArtifactResult,
	DAGRepo repos.DAG,
	integrationRepo repos.Resource,
	operatorRepo repos.Operator,
	storageMigrationRepo repos.StorageMigration,
	DB database.Database,
) error {
	destResourceName := "Local Filesystem"
	var destResourceID *uuid.UUID
	if destResourceObj != nil {
		destResourceName = destResourceObj.Name
		destResourceID = &destResourceObj.ID
	}

	// Begin recording the storage migration lifecycle.
	storageMigrationObj, err := storageMigrationRepo.Create(
		ctx,
		destResourceID,
		DB,
	)
	if err != nil {
		return errors.Wrap(err, "Unable to migrate storage.")
	}

	// If the migration is successful, the new entry is given a success execution status, along with `current=True`.
	// If the migration is unsuccessful, the error is recorded on the new entry in `storage_migration`.
	go func() {
		// Shadows the context in the outer scope on purpose.
		ctx := context.Background()

		log.Info("Starting storage migration process...")
		// Wait until the server is paused
		pauseServer()
		// Makes sure that the server is restarted
		defer restartServer()

		execState := storageMigrationObj.ExecState

		var err error
		defer func() {
			if err != nil {
				execState.UpdateWithFailure(
					// This can be a system error too. But no one cares right now.
					shared.UserFatalFailure,
					&shared.Error{
						Tip:     fmt.Sprintf("Failure occurred when migrating to the new storage integration `%s`.", destResourceName),
						Context: err.Error(),
					},
				)
				err = updateStorageMigrationExecState(ctx, storageMigrationObj.ID, &execState, storageMigrationRepo, DB)
				if err != nil {
					log.Errorf("Unexpected error when updating the storage migration entry to FAILED: %v", err)
					return
				}

			}
		}()

		// Mark the migration explicitly as RUNNING.
		runningAt := time.Now()
		execState.Timestamps.RunningAt = &runningAt
		err = updateStorageMigrationExecState(ctx, storageMigrationObj.ID, &execState, storageMigrationRepo, DB)
		if err != nil {
			log.Errorf("Unexpected error when updating the storage migration entry to RUNNING: %v", err)
			return
		}

		// Actually perform the storage migration.
		// Wait until there are no more workflow runs in progress
		lock := utils.NewExecutionLock()
		if err = lock.Lock(); err != nil {
			err = errors.Wrap(err, "Unexpected error when acquiring workflow execution lock.")
			return
		}
		defer func() {
			if lockErr := lock.Unlock(); lockErr != nil {
				log.Errorf("Unexpected error when unlocking workflow execution lock: %v", lockErr)
			}
		}()

		// Migrate all storage content to the new storage config
		currentStorageConfig := config.Storage()
		storageCleanupConfig, err := MigrateStorageAndVault(
			context.Background(),
			&currentStorageConfig,
			newStorageConfig,
			orgID,
			DAGRepo,
			artifactRepo,
			artifactResultRepo,
			operatorRepo,
			integrationRepo,
			DB,
		)
		// We let the defer() handle the failure case appropriately.
		if err != nil {
			return
		}

		log.Info("Successfully migrated the storage layer!")
		finishedAt := time.Now()
		execState.Timestamps.FinishedAt = &finishedAt
		execState.Status = shared.SucceededExecutionStatus

		// The update of the storage config and storage migration entry should happen together.
		// While we don't enforce this atomically, we can make the two update together to minimize the risk.
		err = updateStorageMigrationExecState(ctx, storageMigrationObj.ID, &execState, storageMigrationRepo, DB)
		if err != nil {
			log.Errorf("Unexpected error when updating the storage migration entry to SUCCESS: %v", err)
			return
		}

		err = config.UpdateStorage(newStorageConfig)
		if err != nil {
			log.Errorf("Unexpected error when updating the global storage layer config: %v", err)
			return
		} else {
			log.Info("Successfully updated the global storage layer config!")
		}

		// We only perform best-effort deletion the old storage layer files here, after everything else has succeede.
		for _, key := range storageCleanupConfig.StoreKeys {
			if err := storageCleanupConfig.Store.Delete(ctx, key); err != nil {
				log.Errorf("Unexpected error when deleting the old storage file %s: %v", key, err)
			}
		}

		for _, key := range storageCleanupConfig.VaultKeys {
			if err := storageCleanupConfig.Vault.Delete(ctx, key); err != nil {
				log.Errorf("Unexpected error when deleting the old vault file %s: %v", key, err)
			}
		}
	}()
	return nil
}

// Also updates `current=True` if the execution state is marked as SUCCESS!
func updateStorageMigrationExecState(
	ctx context.Context,
	storageMigrationID uuid.UUID,
	execState *shared.ExecutionState,
	storageMigrationRepo repos.StorageMigration,
	db database.Database,
) error {
	// If we're updating the old storage migration entry, that must be done in the same transaction
	// as the update of the new storage migration entry, so that there is at most one current=True
	// entry in the storage_migration table.
	txn, err := db.BeginTx(context.Background())
	if err != nil {
		return errors.Wrap(err, "Unable to start transaction for updating storage state.")
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	updates := map[string]interface{}{
		models.StorageMigrationExecutionState: execState,
	}

	// This is updated to a transaction if we also need to mark an old entry as current=False.
	if execState.Status == shared.SucceededExecutionStatus {
		updates[models.StorageMigrationCurrent] = true

		// If there was a previous storage migration, update that entry to be `current=False`.
		oldStorageMigrationObj, err := storageMigrationRepo.Current(ctx, txn)
		if err != nil && !aq_errors.Is(err, database.ErrNoRows()) {
			return errors.Wrap(err, "Unexpected error when fetching current storage state.")
		}
		if err == nil {
			_, err = storageMigrationRepo.Update(
				ctx,
				oldStorageMigrationObj.ID,
				map[string]interface{}{
					models.StorageMigrationCurrent: false,
				},
				txn,
			)
			if err != nil {
				return errors.Wrap(err, "Unexpected error when updating old storage migration entry to be non-current")
			}
		}
		// Continue without doing anything if there was no previous storage migration.
	}

	// Perform the actual intended execution state update.
	_, err = storageMigrationRepo.Update(
		ctx,
		storageMigrationID,
		updates,
		txn,
	)
	if err != nil {
		return errors.Wrap(err, "Unexpected error when updating storage migration execution state.")
	}

	return txn.Commit(ctx)
}

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
	integrationRepo repos.Resource,
	DB database.Database,
) (*StorageCleanupConfig, error) {
	log.Infof("Migrating from %v to %v", *oldConf, *newConf)

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
			return nil, err
		}

		log.Infof("There are %v artifacts to migrate for DAG %v", len(artifacts), dag.ID)

		for _, artifact := range artifacts {
			log.Infof("Starting migration for artifact %v of DAG %v", artifact.ID, dag.ID)

			artifactResults, err := artifactResultRepo.GetByArtifact(ctx, artifact.ID, txn)
			if err != nil {
				return nil, err
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
					return nil, err
				}

				if err == nil {
					// Only try to migrate artifact result if there was no issue reading
					// it from the `oldStore`
					if err := newStore.Put(ctx, artifactResult.ContentPath, val); err != nil {
						log.Errorf("Unable to write artifact result %v to new store: %v", artifactResult.ID, err)
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
				return nil, err
			}

			if err := newStore.Put(ctx, operatorCodePath, val); err != nil {
				log.Errorf("Unable to write operator code %v to new store: %v", operator.ID, err)
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
