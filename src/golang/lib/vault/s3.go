package vault

import (
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/storage"
)

const (
	s3VaultDir = "vault"
)

func newS3Vault(s3StoreConf shared.S3Config, key string) Vault {
	// The S3 vault stores secrets under the ../vault path
	// The S3 bucket is in the form of s3:// so we can't use path.Join, because
	// it will clean the final filepath and change the prefix to s3:/
	// bucket := s3StoreConf.Bucket

	// TODO: fix the messaging
	//if len(bucket) > 0 && bucket[len(bucket)-1] == '/' {
	//	bucket += s3VaultDir
	//} else {
	//	bucket += "/" + s3VaultDir
	//}
	//s3StoreConf.Bucket = bucket
	s3StoreConf.RootDir += "/" + s3VaultDir

	store := storage.NewStorage(&shared.StorageConfig{
		Type:     shared.S3StorageType,
		S3Config: &s3StoreConf,
	})

	return &vault{
		store: store,
		key:   key,
	}
}
