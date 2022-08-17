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
		Val:                po.dbOperator.Spec.Param().Val,
		OutputMetadataPath: po.outputMetadataPaths[0],
		OutputContentPath:  po.outputContentPaths[0],
	}
}

func (po *paramOperatorImpl) Launch(ctx context.Context) error {
	return po.launch(ctx, po.JobSpec())
}
