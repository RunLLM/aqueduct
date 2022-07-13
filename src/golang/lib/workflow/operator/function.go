package operator

import (
	"context"
	"fmt"
	db_artifact "github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/operator/function"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/workflow/scheduler"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

const (
	defaultFunctionEntryPointFile   = "model.py"
	defaultFunctionEntryPointClass  = "Function"
	defaultFunctionEntryPointMethod = "predict"
)

type functionOperatorImpl struct {
	baseOperator
}

func newFunctionOperator(
	ctx context.Context,
	baseFields baseOperator,
) (Operator, error) {
	inputs := baseFields.inputs
	outputs := baseFields.outputs

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

func generateFunctionJobName() string {
	return fmt.Sprintf("function-operator-%s", uuid.New().String())
}

func (fo *functionOperatorImpl) Finish(ctx context.Context) {
	// If the operator was not persisted to the DB, cleanup the serialized function.
	if !fo.resultsPersisted {
		utils.CleanupStorageFile(ctx, fo.storageConfig, fo.dbOperator.Spec.Function().StoragePath)
	}

	fo.baseOperator.Finish(ctx)
}

func (fo *functionOperatorImpl) JobSpec() job.Spec {
	fn := fo.dbOperator.Spec.Function()

	entryPoint := fn.EntryPoint
	if entryPoint == nil {
		entryPoint = &function.EntryPoint{
			File:      defaultFunctionEntryPointFile,
			ClassName: defaultFunctionEntryPointClass,
			Method:    defaultFunctionEntryPointMethod,
		}
	}

	inputArtifactTypes := make([]db_artifact.Type, 0, len(fo.inputs))
	outputArtifactTypes := make([]db_artifact.Type, 0, len(fo.outputs))
	for _, inputArtifact := range fo.inputs {
		inputArtifactTypes = append(inputArtifactTypes, inputArtifact.Type())
	}
	for _, outputArtifact := range fo.outputs {
		outputArtifactTypes = append(outputArtifactTypes, outputArtifact.Type())
	}

	return &job.FunctionSpec{
		BasePythonSpec: job.NewBasePythonSpec(
			job.FunctionJobType,
			fo.jobName,
			*fo.storageConfig,
			fo.opMetadataPath,
		),
		FunctionPath:        fn.StoragePath,
		EntryPointFile:      entryPoint.File,
		EntryPointClass:     entryPoint.ClassName,
		EntryPointMethod:    entryPoint.Method,
		CustomArgs:          fn.CustomArgs,
		InputContentPaths:   fo.inputContentPaths,
		InputMetadataPaths:  fo.inputMetadataPaths,
		OutputContentPaths:  fo.outputContentPaths,
		OutputMetadataPaths: fo.outputMetadataPaths,
		InputArtifactTypes:  inputArtifactTypes,
		OutputArtifactTypes: outputArtifactTypes,
	}
}
