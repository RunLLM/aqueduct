package operator

import (
	"context"
	"fmt"
	db_artifact "github.com/aqueducthq/aqueduct/lib/collections/artifact"
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

	inputs := base.inputs
	outputs := base.outputs

	if len(inputs) != 1 {
		return nil, errWrongNumInputs
	}
	if len(outputs) != 0 {
		return nil, errWrongNumOutputs
	}

	for _, inputArtifact := range inputs {
		if inputArtifact.Type() != db_artifact.TableType {
			return nil, errors.New("Only table artifacts can be saved.")
		}
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

func (lo *loadOperatorImpl) JobSpec() job.Spec {
	spec := lo.dbOperator.Spec.Load()

	return &job.LoadSpec{
		BasePythonSpec: job.NewBasePythonSpec(
			job.LoadJobType,
			lo.jobName,
			*lo.storageConfig,
			lo.opMetadataPath,
		),
		ConnectorName:     spec.Service,
		ConnectorConfig:   lo.config,
		Parameters:        spec.Parameters,
		InputContentPath:  lo.inputContentPaths[0],
		InputMetadataPath: lo.inputMetadataPaths[0],
	}
}
