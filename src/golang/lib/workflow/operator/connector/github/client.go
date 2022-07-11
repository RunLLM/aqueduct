package github

import (
	"context"
	"github.com/aqueducthq/aqueduct/lib/collections/operator/connector"
	"github.com/aqueducthq/aqueduct/lib/collections/operator/function"
)

type Client interface {
	PullAndUpdateFunction(
		ctx context.Context,
		spec *function.Function,
		alwaysPullContent bool,
	) (bool, []byte, error)
	PullExtract(ctx context.Context, spec *connector.Extract) (bool, error)
}
