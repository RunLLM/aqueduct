package dag

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/environment"
)

func (w *workflowDagImpl) FindMissingEnv(
	ctx context.Context,
) ([]environment.Environment, error) {
	return nil, nil
}

func (w *workflowDagImpl) BindOperatorsToEnvs(ctx context.Context) error {
	return nil
}
