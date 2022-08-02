package operator

import (
	"context"
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
	_, inputMetadataPaths := unzipExecPathsToRawPaths(smo.inputExecPaths)

	return &job.SystemMetricSpec{
		BasePythonSpec: job.NewBasePythonSpec(
			job.SystemMetricJobType,
			smo.jobName,
			*smo.storageConfig,
			smo.metadataPath,
		),
		InputMetadataPaths: inputMetadataPaths,
		OutputContentPath:  smo.outputExecPaths[0].ArtifactContentPath,
		OutputMetadataPath: smo.outputExecPaths[0].ArtifactMetadataPath,
		MetricName:         smo.dbOperator.Spec.SystemMetric().MetricName,
	}
}

func (smo *systemMetricOperatorImpl) Launch(ctx context.Context) error {
	return smo.launch(ctx, smo.JobSpec())
}
