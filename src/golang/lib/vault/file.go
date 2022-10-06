package vault

import (
	"context"
	"fmt"
	"os"
)

const (
	FileVaultDir       = "vault/"
	filePermissionCode = 0o664
)

type fileVault struct {
	fileConfig *FileConfig
}

func NewFileVault(conf *FileConfig) (*fileVault, error) {
	return &fileVault{fileConfig: conf}, nil
}

func (f *fileVault) Config() Config {
	return f.fileConfig
}

func (f *fileVault) Put(ctx context.Context, name string, secrets map[string]string) error {
	encrypted, err := encrypt(secrets, f.fileConfig.EncryptionKey)
	if err != nil {
		return err
	}

	return os.WriteFile(f.getFullPath(name), encrypted, filePermissionCode)
}

func (f *fileVault) Get(ctx context.Context, name string) (map[string]string, error) {
	ciphertext, err := os.ReadFile(f.getFullPath(name))
	if err != nil {
		return nil, err
	}

	return decrypt(ciphertext, f.fileConfig.EncryptionKey)
}

func (f *fileVault) Delete(ctx context.Context, name string) error {
	return os.Remove(f.getFullPath(name))
}

func (f *fileVault) getFullPath(key string) string {
	return fmt.Sprintf("%s/%s", f.fileConfig.Directory, key)
}
