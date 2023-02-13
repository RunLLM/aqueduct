package server

import (
	"context"
	"os"
	"path"

	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
)

// syncVaultWithStorage checks if this server's vault is out of sync
// with its storage; if so, it will migrate the vault's contents into
// storage. This operation will only need to happen once.
func syncVaultWithStorage(
	vaultObj vault.Vault,
	integrationRepo repos.Integration,
	DB database.Database,
) error {
	oldVaultPath := path.Join(config.AqueductPath(), "vault")
	if _, err := os.Stat(oldVaultPath); err != nil {
		if os.IsNotExist(err) {
			// The old vault path does not exist, so the vault has already been synced
			// with storage.
			return nil
		}
		return err
	}

	oldVault, err := vault.NewVault(
		&shared.StorageConfig{
			Type: shared.FileStorageType,
			FileConfig: &shared.FileConfig{
				Directory: config.AqueductPath(),
			},
		},
		config.EncryptionKey(),
	)
	if err != nil {
		return err
	}

	if _, err := utils.MigrateVault(
		context.Background(),
		oldVault,
		vaultObj,
		accountOrganizationId,
		integrationRepo,
		DB,
	); err != nil {
		return err
	}

	// Vault syncing was successful so we can delete `oldVaultPath`, so this operation
	// is not repeated.
	return os.RemoveAll(oldVaultPath)
}
