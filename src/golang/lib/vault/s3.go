package vault

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/storage"
)

type s3Vault struct {
	store storage.Storage
}

func (sv *s3Vault) Put(ctx context.Context, name string, secrets map[string]string) error {
	return nil
}

func (sv *s3Vault) Get(ctx context.Context, name string) (map[string]string, error) {
	return nil, nil
}

func (sv *s3Vault) Delete(ctx context.Context, name string) error {
	return nil
}
