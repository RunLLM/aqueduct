package operator

import (
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/workflow/scheduler"
	"github.com/dropbox/godropbox/errors"
)

type checkOperatorImpl struct {
	baseFunctionOperator
}

func newCheckOperator(base baseFunctionOperator) (Operator, error) {
	base.jobName = generateFunctionJobName()

	inputs := base.inputs
	outputs := base.outputs
	if len(inputs) == 0 {
		return nil, scheduler.ErrWrongNumInputs
	}
	if len(outputs) != 1 {
		return nil, scheduler.ErrWrongNumOutputs
	}

	for _, inputArtifact := range inputs {
		if inputArtifact.Type() != artifact.TableType &&
			inputArtifact.Type() != artifact.FloatType &&
			inputArtifact.Type() != artifact.JsonType {
			return nil, errors.New("Inputs to metric operator must be Table, Float, or Parameter Artifacts.")
		}
	}
	for _, outputArtifact := range outputs {
		if outputArtifact.Type() != artifact.BoolType {
			return nil, errors.New("Outputs of function operator must be Table Artifacts.")
		}
	}

	return &checkOperatorImpl{
		base,
	}, nil
}

func (co *checkOperatorImpl) JobSpec() job.Spec {
	fn := co.dbOperator.Spec.Check().Function
	return co.jobSpec(&fn)
}
