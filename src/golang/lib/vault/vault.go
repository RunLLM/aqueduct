package vault

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/dropbox/godropbox/errors"
)

var ErrInvalidVaultConfig = errors.New("Vault config is invalid.")

type Vault interface {
	Put(ctx context.Context, name string, secrets map[string]string) error
	Get(ctx context.Context, name string) (map[string]string, error)
	Delete(ctx context.Context, name string) error
}

// NewVault constructs a Vault from the storage config and encryption key provided.
func NewVault(storageConf *shared.StorageConfig, key string) (Vault, error) {
	switch storageConf.Type {
	case shared.FileStorageType:
		return newFileVault(*storageConf.FileConfig, key)
	default:
		return nil, errors.Newf("Unsupported vault type: %v", storageConf.Type)
	}
}
