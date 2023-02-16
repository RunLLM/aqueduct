package vault

import (
	"path/filepath"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/storage"
)

const (
	FileVaultDir = "vault/"
)

func newFileVault(fileStoreConf shared.FileConfig, key string) Vault {
	// The file vault stores secrets under the ../vault subdirectory
	fileStoreConf.Directory = filepath.Join(fileStoreConf.Directory, FileVaultDir)

	store := storage.NewStorage(&shared.StorageConfig{
		Type:       shared.FileStorageType,
		FileConfig: &fileStoreConf,
	})

	return &vault{
		store: store,
		key:   key,
	}
}
