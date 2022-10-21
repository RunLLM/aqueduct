package vault

import (
	"path"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/storage"
)

const (
	s3VaultDir = "vault"
)

func newS3Vault(s3StoreConf shared.S3Config, key string) Vault {
	// The S3 vault stores secrets under the ../vault path
	s3StoreConf.Bucket = path.Join(s3StoreConf.Bucket, s3VaultDir)

	store := storage.NewStorage(&shared.StorageConfig{
		Type:     shared.S3StorageType,
		S3Config: &s3StoreConf,
	})

	return &vault{
		store: store,
		key:   key,
	}
}
