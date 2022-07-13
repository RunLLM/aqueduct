package operator

import (
	"fmt"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/google/uuid"
)

func generateSystemMetricJobName() string {
	return fmt.Sprintf("system-metric-operator-%s", uuid.New().String())
}

type systemMetricOperatorImpl struct {
	baseOperator
}

func newSystemMetricOperator(
	base baseOperator,
) (Operator, error) {
	base.jobName = generateSystemMetricJobName()

	inputs := base.inputs
	outputs := base.outputs

	if len(inputs) != 1 {
		return nil, errWrongNumInputs
	}
	if len(outputs) != 1 {
		return nil, errWrongNumOutputs
	}

	return &systemMetricOperatorImpl{
		base,
	}, nil
}

func (smo *systemMetricOperatorImpl) JobSpec() job.Spec {
	return &job.SystemMetricSpec{
		BasePythonSpec: job.NewBasePythonSpec(
			job.SystemMetricJobType,
			smo.jobName,
			*smo.storageConfig,
			smo.opMetadataPath,
		),
		InputMetadataPaths: smo.inputMetadataPaths,
		OutputContentPath:  smo.outputContentPaths[0],
		OutputMetadataPath: smo.outputMetadataPaths[0],
		MetricName:         smo.dbOperator.Spec.SystemMetric().MetricName,
	}
}
