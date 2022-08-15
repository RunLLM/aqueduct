package operator

import (
	"context"

	db_artifact "github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/job"
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
		return nil, errWrongNumInputs
	}
	if len(outputs) != 1 {
		return nil, errWrongNumOutputs
	}

	for _, inputArtifact := range inputs {
		if inputArtifact.Type() != db_artifact.TableType &&
			inputArtifact.Type() != db_artifact.FloatType &&
			inputArtifact.Type() != db_artifact.JsonType {
			return nil, errors.New("Inputs to metric operator must be Table, Float, or Parameter Artifacts.")
		}
	}
	for _, outputArtifact := range outputs {
		if outputArtifact.Type() != db_artifact.BoolType {
			return nil, errors.New("Outputs of function operator must be Table Artifacts.")
		}
	}

	return &checkOperatorImpl{
		base,
	}, nil
}

func (co *checkOperatorImpl) JobSpec() job.Spec {
	fn := co.dbOperator.Spec.Check().Function
	return co.jobSpec(&fn, &co.dbOperator.Spec.Check().Level)
}

func (co *checkOperatorImpl) Launch(ctx context.Context) error {
	return co.launch(ctx, co.JobSpec())
}
