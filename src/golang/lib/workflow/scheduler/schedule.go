package scheduler

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
)

var (
	ErrWrongNumInputs                = errors.New("Wrong number of operator inputs")
	ErrWrongNumOutputs               = errors.New("Wrong number of operator outputs")
	ErrWrongNumMetadataInputs        = errors.New("Wrong number of input metadata paths for operator")
	ErrWrongNumArtifactContentPaths  = errors.New("Wrong number of artifact content paths.")
	ErrWrongNumArtifactMetadataPaths = errors.New("Wrong number of artifact metadata paths.")
)

// ScheduleOperator executes an operator based on its spec.
// Inputs:
//	op: the operator to execute
//	inputs: a list of input artifacts
//	outputs: a list of output artifacts
//	artifactPaths: a pre-generated map of `artifactId -> storage paths`. It must cover all artifacts in the workflow
//	operatorMetadataPath: a pre-generated storage path to store the intermediate operator metadata.
//
// Outputs:
//	string: cron-job ID to track operator execution status.
//	error: any error that the caller should handle.
//
// Does:
// 	This function calls the corresponding executor based on the type of `spec`. It deserialize the log in
// 	the storage path as a part of the returned results. It also updates the `OperatorResult` in data model.
//
// Assumptions:
//	We use a switch to call type-specific executors for each operator type. Each type-specific executor should:
//	- Ideally managed in its own .go file under operator_execution package
//	- return (error)
//	- serialize operator metadata to `operatorMetadataPath` in json format
//	- properly deserialize input artifacts and output artifacts, based on `inputs`, `outputs`, and `artifactPaths`.
//
func ScheduleOperator(
	ctx context.Context,
	op operator.DBOperator,
	inputArtifacts []artifact.DBArtifact,
	outputArtifacts []artifact.DBArtifact,
	metadataPath string,
	inputContentPaths []string,
	inputMetadataPaths []string,
	outputContentPaths []string,
	outputMetadataPaths []string,
	storageConfig *shared.StorageConfig,
	jobManager job.JobManager,
	vaultObject vault.Vault,
) (string, error) {
	// Append to this switch for newly supported operator types
	if op.Spec.IsFunction() {
		// A function operator takes any number of dataframes as input and outputs
		// any number of dataframes.
		inputArtifactTypes := make([]artifact.Type, 0, len(inputArtifacts))
		for _, inputArtifact := range inputArtifacts {
			if inputArtifact.Spec.Type() != artifact.TableType && inputArtifact.Spec.Type() != artifact.JsonType {
				return "", errors.New("Inputs to function operator must be Table or Parameter Artifacts.")
			}
			inputArtifactTypes = append(inputArtifactTypes, inputArtifact.Spec.Type())
		}
		outputArtifactTypes := make([]artifact.Type, 0, len(outputArtifacts))
		for _, outputArtifact := range outputArtifacts {
			if outputArtifact.Spec.Type() != artifact.TableType {
				return "", errors.New("Outputs of function operator must be Table Artifacts.")
			}
			outputArtifactTypes = append(outputArtifactTypes, outputArtifact.Spec.Type())
		}

		return ScheduleFunction(
			ctx,
			*op.Spec.Function(),
			metadataPath,
			inputContentPaths,
			inputMetadataPaths,
			outputContentPaths,
			outputMetadataPaths,
			inputArtifactTypes,
			outputArtifactTypes,
			storageConfig,
			jobManager,
		)
	}

	if op.Spec.IsMetric() {
		if len(outputArtifacts) != 1 {
			return "", ErrWrongNumOutputs
		}

		inputArtifactTypes := make([]artifact.Type, 0, len(inputArtifacts))
		for _, inputArtifact := range inputArtifacts {
			if inputArtifact.Spec.Type() != artifact.TableType &&
				inputArtifact.Spec.Type() != artifact.FloatType &&
				inputArtifact.Spec.Type() != artifact.JsonType {
				return "", errors.New("Inputs to metric operator must be Table, Float, or Parameter Artifacts.")
			}
			inputArtifactTypes = append(inputArtifactTypes, inputArtifact.Spec.Type())
		}
		outputArtifactTypes := []artifact.Type{artifact.FloatType}

		return ScheduleFunction(
			ctx,
			op.Spec.Metric().Function,
			metadataPath,
			inputContentPaths,
			inputMetadataPaths,
			outputContentPaths,
			outputMetadataPaths,
			inputArtifactTypes,
			outputArtifactTypes,
			storageConfig,
			jobManager,
		)
	}

	if op.Spec.IsCheck() {
		if len(outputArtifacts) != 1 {
			return "", ErrWrongNumOutputs
		}

		// Checks can be computed on tables and metrics.
		inputArtifactTypes := make([]artifact.Type, 0, len(inputArtifacts))
		for _, inputArtifact := range inputArtifacts {
			if inputArtifact.Spec.Type() != artifact.TableType &&
				inputArtifact.Spec.Type() != artifact.FloatType &&
				inputArtifact.Spec.Type() != artifact.JsonType {
				return "", errors.New("Inputs to metric operator must be Table, Float, or Parameter Artifacts.")
			}
			inputArtifactTypes = append(inputArtifactTypes, inputArtifact.Spec.Type())
		}
		outputArtifactTypes := []artifact.Type{artifact.BoolType}

		return ScheduleFunction(
			ctx,
			op.Spec.Check().Function,
			metadataPath,
			inputContentPaths,
			inputMetadataPaths,
			outputContentPaths,
			outputMetadataPaths,
			inputArtifactTypes,
			outputArtifactTypes,
			storageConfig,
			jobManager,
		)
	}

	if op.Spec.IsExtract() {
		inputParamNames := make([]string, 0, len(inputArtifacts))
		for _, inputArtifact := range inputArtifacts {
			if inputArtifact.Spec.Type() != artifact.JsonType {
				return "", errors.New("Only parameters can be used as inputs to extract operators.")
			}
			inputParamNames = append(inputParamNames, inputArtifact.Name)
		}

		if len(outputArtifacts) != 1 {
			return "", ErrWrongNumOutputs
		}
		if len(outputContentPaths) != 1 {
			return "", ErrWrongNumArtifactContentPaths
		}
		if len(outputMetadataPaths) != 1 {
			return "", ErrWrongNumArtifactMetadataPaths
		}

		return ScheduleExtract(
			ctx,
			*op.Spec.Extract(),
			metadataPath,
			inputParamNames,
			inputContentPaths,
			inputMetadataPaths,
			outputContentPaths[0],
			outputMetadataPaths[0],
			storageConfig,
			jobManager,
			vaultObject,
		)
	}

	if op.Spec.IsLoad() {
		if len(inputArtifacts) != 1 {
			return "", ErrWrongNumInputs
		}
		if len(outputArtifacts) != 0 {
			return "", ErrWrongNumOutputs
		}
		if len(inputContentPaths) != 1 {
			return "", ErrWrongNumArtifactContentPaths
		}
		if len(inputMetadataPaths) != 1 {
			return "", ErrWrongNumArtifactMetadataPaths
		}
		return ScheduleLoad(
			ctx,
			*op.Spec.Load(),
			metadataPath,
			inputContentPaths[0],
			inputMetadataPaths[0],
			storageConfig,
			jobManager,
			vaultObject,
		)
	}

	if op.Spec.IsParam() {
		if len(inputArtifacts) != 0 {
			return "", ErrWrongNumInputs
		}
		if len(outputArtifacts) != 1 {
			return "", ErrWrongNumOutputs
		}
		if !outputArtifacts[0].Spec.IsJson() {
			return "", errors.Newf("Internal Error: parameter must output a JSON artifact.")
		}
		if len(outputContentPaths) != 1 {
			return "", ErrWrongNumArtifactContentPaths
		}
		if len(outputMetadataPaths) != 1 {
			return "", ErrWrongNumArtifactMetadataPaths
		}

		return ScheduleParam(
			ctx,
			*op.Spec.Param(),
			metadataPath,
			outputContentPaths[0],
			outputMetadataPaths[0],
			storageConfig,
			jobManager,
		)
	}

	if op.Spec.IsSystemMetric() {
		if len(outputContentPaths) != 1 {
			return "", ErrWrongNumArtifactContentPaths
		}
		if len(outputMetadataPaths) != 1 {
			return "", ErrWrongNumArtifactMetadataPaths
		}
		// We currently allow the spec to contain multiple input_metadata paths.
		// A system metric currently spans over a single operator, so we enforce that here
		if len(inputMetadataPaths) != 1 {
			return "", ErrWrongNumMetadataInputs
		}

		return ScheduleSystemMetric(
			ctx,
			*op.Spec.SystemMetric(),
			metadataPath,
			inputMetadataPaths,
			outputContentPaths[0],
			outputMetadataPaths[0],
			storageConfig,
			jobManager,
		)
	}

	// If we reach here, the operator opSpec type is not supported.
	return "", errors.Newf("Unsupported operator opSpec with type %s", op.Spec.Type())
}

// CheckOperatorExecutionStatus returns the operator metadata (if it exists) and the operator status
// of a completed job.
func CheckOperatorExecutionStatus(
	ctx context.Context,
	storageConfig *shared.StorageConfig,
	operatorMetadataPath string,
) *shared.ExecutionState {
	var logs shared.ExecutionState
	err := utils.ReadFromStorage(
		ctx,
		storageConfig,
		operatorMetadataPath,
		&logs,
	)
	if err != nil {
		// Treat this as a system internal error since operator metadata was not found
		log.Errorf(
			"Unable to read operator metadata from storage. Operator may have failed before writing metadata. %v",
			err,
		)

		failureType := shared.SystemFailure
		return &shared.ExecutionState{
			Status:      shared.FailedExecutionStatus,
			FailureType: &failureType,
			Error: &shared.Error{
				Context: fmt.Sprintf("%v", err),
				Tip:     shared.TipUnknownInternalError,
			},
		}
	}

	return &logs
}
