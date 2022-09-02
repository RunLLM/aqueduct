package operator

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/google/uuid"
)

type functionOperatorImpl struct {
	baseFunctionOperator
}

func generateFunctionJobName() string {
	return fmt.Sprintf("function-operator-%s", uuid.New().String())
}

func newFunctionOperator(
	base baseFunctionOperator,
) (Operator, error) {
	base.jobName = generateFunctionJobName()

	return &functionOperatorImpl{
		base,
	}, nil
}

func (fo *functionOperatorImpl) JobSpec() job.Spec {
	fn := fo.dbOperator.Spec.Function()
	return fo.jobSpec(fn, nil /* checkSeverity */)
}

func (fo *functionOperatorImpl) Launch(ctx context.Context) error {
	return fo.launch(ctx, fo.JobSpec())
}
