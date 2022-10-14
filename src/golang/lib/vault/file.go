package vault

import (
	"context"
	"path/filepath"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/storage"
)

const (
	FileVaultDir = "vault/"
)

type fileVault struct {
	store storage.Storage
	key   string
}

func newFileVault(fileStoreConf shared.FileConfig, key string) (Vault, error) {
	// The file vault stores secrets under the ../vault subdirectory
	fileStoreConf.Directory = filepath.Join(fileStoreConf.Directory, FileVaultDir)

	store := storage.NewStorage(&shared.StorageConfig{
		Type:       shared.FileStorageType,
		FileConfig: &fileStoreConf,
	})

	return &fileVault{
		store: store,
		key:   key,
	}, nil
}

func (f *fileVault) Put(ctx context.Context, name string, secrets map[string]string) error {
	encrypted, err := encrypt(secrets, f.key)
	if err != nil {
		return err
	}

	return f.store.Put(ctx, name, encrypted)
}

func (f *fileVault) Get(ctx context.Context, name string) (map[string]string, error) {
	ciphertext, err := f.store.Get(ctx, name)
	if err != nil {
		return nil, err
	}

	return decrypt(ciphertext, f.key)
}

func (f *fileVault) Delete(ctx context.Context, name string) error {
	return f.store.Delete(ctx, name)
}
