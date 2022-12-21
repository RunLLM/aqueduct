package operator

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
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
		if inputArtifact.Type() != shared.StringArtifact {
			return nil, errors.New("Only strings can be used as inputs to extract operators.")
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

func (eo *extractOperatorImpl) JobSpec() (returnedSpec job.Spec) {
	spec := eo.dbOperator.Spec.Extract()

	inputParamNames := make([]string, 0, len(eo.inputs))
	for _, inputArtifact := range eo.inputs {
		inputParamNames = append(inputParamNames, inputArtifact.Name())
	}

	inputContentPaths, inputMetadataPaths := unzipExecPathsToRawPaths(eo.inputExecPaths)
	outputContentPaths, outputMetadataPaths := unzipExecPathsToRawPaths(eo.outputExecPaths)

	return &job.ExtractSpec{
		BasePythonSpec: job.NewBasePythonSpec(
			job.ExtractJobType,
			eo.jobName,
			*eo.storageConfig,
			eo.metadataPath,
		),
		InputParamNames:    inputParamNames,
		InputContentPaths:  inputContentPaths,
		InputMetadataPaths: inputMetadataPaths,
		ConnectorName:      spec.Service,
		ConnectorConfig:    eo.config,
		Parameters:         spec.Parameters,
		OutputContentPath:  outputContentPaths[0],
		OutputMetadataPath: outputMetadataPaths[0],
	}
}

func (eo *extractOperatorImpl) Launch(ctx context.Context) error {
	return eo.launch(ctx, eo.JobSpec())
}
