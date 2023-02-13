package vault

import (
	"path"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/storage"
)

const (
	gcsVaultDir = "vault"
)

func newGCSVault(gcsStoreConf shared.GCSConfig, key string) Vault {
	// The GCS vault stores secrets under the ../vault path
	gcsStoreConf.Bucket = path.Join(gcsStoreConf.Bucket, gcsVaultDir)

	store := storage.NewStorage(&shared.StorageConfig{
		Type:      shared.GCSStorageType,
		GCSConfig: &gcsStoreConf,
	})

	return &vault{
		store: store,
		key:   key,
	}
}
