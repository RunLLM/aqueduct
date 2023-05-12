package operator

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

func generateLoadJobName() string {
	return fmt.Sprintf("load-operator-%s", uuid.New().String())
}

type loadOperatorImpl struct {
	baseOperator

	config auth.Config
}

func newLoadOperator(
	ctx context.Context,
	base baseOperator,
) (Operator, error) {
	base.jobName = generateLoadJobName()

	if base.previewCacheManager != nil {
		return nil, errors.Newf("A load operator cannot be part of a cache-aware workflow execution, since it is non-preview only.")
	}

	inputs := base.inputs
	outputs := base.outputs

	if len(inputs) != 1 {
		return nil, errWrongNumInputs
	}
	if len(outputs) != 0 {
		return nil, errWrongNumOutputs
	}

	spec := base.dbOperator.Spec.Load()
	config, err := auth.ReadConfigFromSecret(ctx, spec.IntegrationId, base.vaultObject)
	if err != nil {
		return nil, err
	}

	return &loadOperatorImpl{
		baseOperator: base,
		config:       config,
	}, nil
}

func (lo *loadOperatorImpl) JobSpec() (returnedSpec job.Spec) {
	spec := lo.dbOperator.Spec.Load()

	inputContentPaths, inputMetadataPaths := unzipExecPathsToRawPaths(lo.inputExecPaths)

	return &job.LoadSpec{
		BasePythonSpec: job.NewBasePythonSpec(
			job.LoadJobType,
			lo.jobName,
			*lo.storageConfig,
			lo.metadataPath,
		),
		ConnectorName:      spec.Service,
		ConnectorConfig:    lo.config,
		Parameters:         spec.Parameters,
		InputContentPaths:  inputContentPaths,
		InputMetadataPaths: inputMetadataPaths,
	}
}

func (lo *loadOperatorImpl) Launch(ctx context.Context) error {
	return lo.launch(ctx, lo.JobSpec())
}
