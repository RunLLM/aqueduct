package github

import (
	"context"

	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

var ErrInvalidManagerConfig = errors.New("Invalid github manager config.")

type Manager interface {
	Config() ManagerConfig
	GetClient(ctx context.Context, userId uuid.UUID) (Client, error)
}

func NewManager(config ManagerConfig) (Manager, error) {
	if config.Type() == NoopManagerType {
		return NewUnimplementedManager(), nil
	}

	return nil, ErrInvalidManagerConfig
}
