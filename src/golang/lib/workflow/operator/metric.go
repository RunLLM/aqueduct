package operator

import (
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/dropbox/godropbox/errors"
)

type metricOperatorImpl struct {
	baseFunctionOperator
}

func newMetricOperator(base baseFunctionOperator) (Operator, error) {
	base.jobName = generateFunctionJobName()

	inputs := base.inputs
	outputs := base.outputs
	if len(inputs) == 0 {
		return nil, errWrongNumInputs
	}
	if len(outputs) != 1 {
		return nil, errWrongNumOutputs
	}

	for _, inputArtifact := range inputs {
		if inputArtifact.Type() != artifact.TableType &&
			inputArtifact.Type() != artifact.FloatType &&
			inputArtifact.Type() != artifact.JsonType {
			return nil, errors.New("Inputs to metric operator must be Table, Float, or Parameter Artifacts.")
		}
	}
	for _, outputArtifact := range outputs {
		if outputArtifact.Type() != artifact.FloatType {
			return nil, errors.New("Outputs of function operator must be Table Artifacts.")
		}
	}

	return &metricOperatorImpl{
		base,
	}, nil
}

func (mo *metricOperatorImpl) JobSpec() job.Spec {
	fn := mo.dbOperator.Spec.Metric().Function
	return mo.jobSpec(&fn)
}
