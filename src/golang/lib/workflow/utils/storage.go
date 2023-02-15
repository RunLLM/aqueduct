package utils

import (
	"context"
	"encoding/json"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
)

func CleanupStorageFile(ctx context.Context, storageConfig *shared.StorageConfig, key string) {
	CleanupStorageFiles(ctx, storageConfig, []string{key})
}

func CleanupStorageFiles(ctx context.Context, storageConfig *shared.StorageConfig, keys []string) {
	for _, key := range keys {
		err := storage.NewStorage(storageConfig).Delete(ctx, key)
		if err != nil {
			log.Errorf("Unable to clean up storage file with key: %s. %v. \n %s", key, err, errors.New("").GetStack())
		}
	}
}

func ObjectExistsInStorage(ctx context.Context, storageConfig *shared.StorageConfig, path string) bool {
	return storage.NewStorage(storageConfig).Exists(ctx, path)
}

func ReadFromStorage(ctx context.Context, storageConfig *shared.StorageConfig, path string, container interface{}) error {
	// Read data from storage and deserialize payload to `container`
	serializedPayload, err := storage.NewStorage(storageConfig).Get(ctx, path)
	if err != nil {
		return errors.Wrap(err, "Unable to get object from storage")
	}

	err = json.Unmarshal(serializedPayload, container)
	if err != nil {
		return errors.Wrap(err, "Unable to unmarshal json payload to container")
	}

	return nil
}
