package github

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/function"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

var ErrGithubNotImplemented = errors.New("Github features are not implemented yet.")

type UnimplementedManager struct{}

func NewUnimplementedManager() *UnimplementedManager {
	return &UnimplementedManager{}
}

func (*UnimplementedManager) Config() ManagerConfig {
	return &UnimplementedManagerConfig{}
}

func (*UnimplementedManager) GetClient(ctx context.Context, userId uuid.UUID) (Client, error) {
	return &UnimplementedClient{}, nil
}

type UnimplementedClient struct{}

func (*UnimplementedClient) PullAndUpdateFunction(ctx context.Context, spec *function.Function, alwaysExtract bool) (bool, []byte, error) {
	return false, nil, ErrGithubNotImplemented
}

func (*UnimplementedClient) PullExtract(ctx context.Context, spec *connector.Extract) (bool, error) {
	return false, ErrGithubNotImplemented
}
