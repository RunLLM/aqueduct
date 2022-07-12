package operator

import (
	db_artifact "github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/aqueducthq/aqueduct/lib/workflow/scheduler"
	"github.com/dropbox/godropbox/errors"
)

type functionOperatorImpl struct {
	baseOperatorFields
}

func newFunctionOperator(
	baseFields baseOperatorFields,
	inputs []artifact.Artifact,
	outputs []artifact.Artifact,
) (Operator, error) {
	if len(inputs) == 0 {
		return nil, scheduler.ErrWrongNumInputs
	}
	if len(outputs) == 0 {
		return nil, scheduler.ErrWrongNumOutputs
	}

	for _, inputArtifact := range inputs {
		if inputArtifact.Type() != db_artifact.TableType && inputArtifact.Type() != db_artifact.JsonType {
			return nil, errors.New("Inputs to function operator must be Table or Parameter Artifacts.")
		}
	}
	for _, outputArtifact := range outputs {
		if outputArtifact.Type() != db_artifact.TableType {
			return nil, errors.New("Outputs of function operator must be Table Artifacts.")
		}
	}

	return &functionOperatorImpl{
		baseFields,
	}, nil
}

func (fo *functionOperatorImpl) Schedule() error {
	if !fo.Ready() {
		return errors.Newf("Operator %s cannot be scheduled yet becuase it isn't ready.", fo.Name())
	}
	if len(fo.jobName) > 0 {
		return errors.Newf("Operator %s was scheduled twice.", fo.Name())
	}

	inputArtifactTypes := make([]db_artifact.Type, 0, len(fo.inputs))
	outputArtifactTypes := make([]db_artifact.Type, 0, len(fo.outputs))
	for _, inputArtifact := range fo.inputs {
		inputArtifactTypes = append(inputArtifactTypes, inputArtifact.Type())
	}
	for _, outputArtifact := range fo.outputs {
		outputArtifactTypes = append(outputArtifactTypes, outputArtifact.Type())
	}

	jobName, err := scheduler.ScheduleFunction(
		fo.ctx,
		*fo.dbOperator.Spec.Function(),
		fo.opMetadataPath,
		fo.inputContentPaths,
		fo.inputMetadataPaths,
		fo.outputContentPaths,
		fo.outputMetadataPaths,
		inputArtifactTypes,
		outputArtifactTypes,
		fo.storageConfig,
		fo.jobManager,
	)
	if err != nil {
		return err
	}

	fo.jobName = jobName
	return nil
}
