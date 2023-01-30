package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
)

const (
	filePermissionCode = 0o664
	dirPermissionCode  = 0o777
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
		return nil, errors.New("Object does not exist in storage.")
	}
	return os.ReadFile(path)
}

func (f *fileStorage) Put(ctx context.Context, key string, value []byte) error {
	filePath := f.getFullPath(key)
	dir := path.Dir(filePath)

	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		// Directory does not exist, so we need to create it
		if err := os.MkdirAll(dir, dirPermissionCode); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return os.WriteFile(filePath, value, filePermissionCode)
}

func (f *fileStorage) Delete(ctx context.Context, key string) error {
	return os.Remove(f.getFullPath(key))
}

func (f *fileStorage) getFullPath(key string) string {
	return fmt.Sprintf("%s/%s", f.fileConfig.Directory, key)
}
