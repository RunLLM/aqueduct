package operator

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/google/uuid"
)

func generateParamJobName() string {
	return fmt.Sprintf("param-operator-%s", uuid.New().String())
}

type paramOperatorImpl struct {
	baseOperator
}

func newParamOperator(
	base baseOperator,
) (Operator, error) {
	base.jobName = generateParamJobName()

	inputs := base.inputs
	outputs := base.outputs
	if len(inputs) != 0 {
		return nil, errWrongNumInputs
	}
	if len(outputs) != 1 {
		return nil, errWrongNumOutputs
	}

	return &paramOperatorImpl{
		base,
	}, nil
}

func (po *paramOperatorImpl) JobSpec() job.Spec {
	return &job.ParamSpec{
		BasePythonSpec: job.NewBasePythonSpec(
			job.ParamJobType,
			po.jobName,
			*po.storageConfig,
			po.metadataPath,
		),
		Val:               po.dbOperator.Spec.Param().Val,
		ExpectedType:      po.outputs[0].Type(),
		SerializationType: po.dbOperator.Spec.Param().SerializationType,

		OutputMetadataPath: po.outputExecPaths[0].ArtifactMetadataPath,
		OutputContentPath:  po.outputExecPaths[0].ArtifactContentPath,
	}
}

// All the parameter operator does is write the parameter value to storage,
// to be consuemd by downstream operators.
func (po *paramOperatorImpl) Launch(ctx context.Context) error {
	return po.launch(ctx, po.JobSpec())
}
