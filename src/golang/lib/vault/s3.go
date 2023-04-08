package vault

import (
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/storage"
)

const (
	s3VaultDir = "vault"
)

func newS3Vault(s3StoreConf shared.S3Config, key string) Vault {
	// The S3 vault stores secrets under the [root_dir]/vault path
	// NOTE: The existing root directory is expected to always end with a slash.
	s3StoreConf.RootDir += s3VaultDir + "/"

	store := storage.NewStorage(&shared.StorageConfig{
		Type:     shared.S3StorageType,
		S3Config: &s3StoreConf,
	})

	return &vault{
		store: store,
		key:   key,
	}
}
