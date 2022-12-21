package vault

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/dropbox/godropbox/errors"
)

type Vault interface {
	Put(ctx context.Context, name string, secrets map[string]string) error
	Get(ctx context.Context, name string) (map[string]string, error)
	Delete(ctx context.Context, name string) error
}

// NewVault constructs a Vault from the storage config and encryption key provided.
func NewVault(storageConf *shared.StorageConfig, key string) (Vault, error) {
	switch storageConf.Type {
	case shared.FileStorageType:
		return newFileVault(*storageConf.FileConfig, key), nil
	case shared.S3StorageType:
		return newS3Vault(*storageConf.S3Config, key), nil
	case shared.GCSStorageType:
		return newGCSVault(*storageConf.GCSConfig, key), nil
	default:
		return nil, errors.Newf("Unsupported vault type: %v", storageConf.Type)
	}
}

type vault struct {
	store storage.Storage
	key   string
}

func (v *vault) Put(ctx context.Context, name string, secrets map[string]string) error {
	encrypted, err := encrypt(secrets, v.key)
	if err != nil {
		return err
	}

	return v.store.Put(ctx, name, encrypted)
}

func (v *vault) Get(ctx context.Context, name string) (map[string]string, error) {
	ciphertext, err := v.store.Get(ctx, name)
	if err != nil {
		return nil, err
	}

	return decrypt(ciphertext, v.key)
}

func (v *vault) Delete(ctx context.Context, name string) error {
	return v.store.Delete(ctx, name)
}
