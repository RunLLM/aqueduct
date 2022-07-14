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

func generateExtractJobName() string {
	return fmt.Sprintf("extract-operator-%s", uuid.New().String())
}

type extractOperatorImpl struct {
	baseOperator

	config auth.Config
}

func newExtractOperator(
	ctx context.Context,
	base baseOperator,
) (Operator, error) {
	base.jobName = generateExtractJobName()

	inputs := base.inputs
	outputs := base.outputs

	if len(outputs) != 1 {
		return nil, errWrongNumOutputs
	}

	for _, inputArtifact := range inputs {
		if inputArtifact.Type() != db_artifact.JsonType {
			return nil, errors.New("Only parameters can be used as inputs to extract operators.")
		}
	}

	spec := base.dbOperator.Spec.Extract()
	config, err := auth.ReadConfigFromSecret(ctx, spec.IntegrationId, base.vaultObject)
	if err != nil {
		return nil, err
	}

	return &extractOperatorImpl{
		baseOperator: base,
		config:       config,
	}, nil
}

func (eo *extractOperatorImpl) JobSpec() job.Spec {
	spec := eo.dbOperator.Spec.Extract()

	inputParamNames := make([]string, 0, len(eo.inputs))
	for _, inputArtifact := range eo.inputs {
		inputParamNames = append(inputParamNames, inputArtifact.Name())
	}

	return &job.ExtractSpec{
		BasePythonSpec: job.NewBasePythonSpec(
			job.ExtractJobType,
			eo.jobName,
			*eo.storageConfig,
			eo.metadataPath,
		),
		InputParamNames:    inputParamNames,
		InputContentPaths:  eo.inputContentPaths,
		InputMetadataPaths: eo.inputMetadataPaths,
		ConnectorName:      spec.Service,
		ConnectorConfig:    eo.config,
		Parameters:         spec.Parameters,
		OutputContentPath:  eo.outputContentPaths[0],
		OutputMetadataPath: eo.outputMetadataPaths[0],
	}
}
