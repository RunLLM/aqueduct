package operator

import (
	"github.com/aqueducthq/aqueduct/lib/job"
)

type checkOperatorImpl struct {
	baseFunctionOperator
}

func newCheckOperator(base baseFunctionOperator) (Operator, error) {
	base.jobName = generateFunctionJobName()

	inputs := base.inputs
	outputs := base.outputs
	if len(inputs) == 0 {
		return nil, errWrongNumInputs
	}
	if len(outputs) != 1 {
		return nil, errWrongNumOutputs
	}

	return &checkOperatorImpl{
		base,
	}, nil
}

func (co *checkOperatorImpl) JobSpec() job.Spec {
	fn := co.dbOperator.Spec.Check().Function
	return co.jobSpec(&fn, &co.dbOperator.Spec.Check().Level)
}
