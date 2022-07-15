package storage

import (
	"context"
	"github.com/dropbox/godropbox/errors"
	"log"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
)

var ErrObjectDoesNotExist = errors.New("Object does not exist in storage.")

type Storage interface {

	// Throws `ErrObjectDoesNotExist` if the path does not exist.
	// TODO(..): this is not the case for S3.
	Get(ctx context.Context, key string) ([]byte, error)
	Put(ctx context.Context, key string, value []byte) error
	Delete(ctx context.Context, key string) error
}

func NewStorage(config *shared.StorageConfig) Storage {
	if config == nil {
		log.Fatalf("Nil storage config.")
	}

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
