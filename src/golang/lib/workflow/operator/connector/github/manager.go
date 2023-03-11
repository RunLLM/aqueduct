package github

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/google/uuid"
)

type Manager interface {
	Config() ManagerConfig
	GetClient(ctx context.Context, userId uuid.UUID) (Client, error)
}

func NewManager(config ManagerConfig) (Manager, error) {
	if config.Type() == NoopManagerType {
		return NewUnimplementedManager(), nil
	}

	return nil, errors.New("Invalid github manager config.")
}
