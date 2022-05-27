package storage

import (
	"context"
	"log"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
)

type Storage interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Put(ctx context.Context, key string, value []byte) error
	Delete(ctx context.Context, key string) error
}

func NewStorage(config *shared.StorageConfig) Storage {
	switch config.Type {
	case shared.S3StorageType:
		return newS3Storage(config.S3Config)
	case shared.FileStorageType:
		return newFileStorage(config.FileConfig)
	default:
		log.Fatalf("Unsupported storage type: %s", config.Type)
		return nil
	}
}
