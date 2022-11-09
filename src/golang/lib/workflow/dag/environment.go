package dag

import (
	"context"

	execEnv "github.com/aqueducthq/aqueduct/lib/execution_environment"
)

func (w *workflowDagImpl) FindMissingExecEnv(
	ctx context.Context,
) ([]execEnv.ExecutionEnvironment, error) {
	return nil, nil
}

func (w *workflowDagImpl) BindOperatorsToEnvs(ctx context.Context) error {
	return nil
}
