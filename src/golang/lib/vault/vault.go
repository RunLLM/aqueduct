package vault

import (
	"context"

	"github.com/dropbox/godropbox/errors"
)

var ErrInvalidVaultConfig = errors.New("Vault config is invalid.")

type Vault interface {
	Config() Config
	Put(ctx context.Context, name string, secrets map[string]string) error
	Get(ctx context.Context, name string) (map[string]string, error)
	Delete(ctx context.Context, name string) error
}

func NewVault(conf Config) (Vault, error) {
	if conf.Type() == FileType {
		fileConfig, ok := conf.(*FileConfig)
		if !ok {
			return nil, ErrInvalidVaultConfig
		}
		return NewFileVault(fileConfig)
	}

	return nil, ErrInvalidVaultConfig
}
