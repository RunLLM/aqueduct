package storage

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
)

const (
	filePermissionCode = 0o664
)

type fileStorage struct {
	fileConfig *shared.FileConfig
}

func newFileStorage(fileConfig *shared.FileConfig) *fileStorage {
	return &fileStorage{
		fileConfig: fileConfig,
	}
}

func (f *fileStorage) Get(ctx context.Context, key string) ([]byte, error) {
	path := f.getFullPath(key)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, ErrObjectDoesNotExist
	}
	return os.ReadFile(path)
}

func (f *fileStorage) Put(ctx context.Context, key string, value []byte) error {
	return os.WriteFile(f.getFullPath(key), value, filePermissionCode)
}

func (f *fileStorage) Delete(ctx context.Context, key string) error {
	return os.Remove(f.getFullPath(key))
}

func (f *fileStorage) getFullPath(key string) string {
	return fmt.Sprintf("%s/%s", f.fileConfig.Directory, key)
}
