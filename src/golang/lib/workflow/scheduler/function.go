package scheduler

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/function"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

const (
	defaultFunctionEntryPointFile   = "model.py"
	defaultFunctionEntryPointClass  = "Function"
	defaultFunctionEntryPointMethod = "predict"
)

func generateFunctionJobName() string {
	return fmt.Sprintf("function-operator-%s", uuid.New().String())
}

func ScheduleFunction(
	ctx context.Context,
	fn function.Function,
	metadataPath string,
	inputContentPaths []string,
	inputMetadataPaths []string,
	outputContentPaths []string,
	outputMetadataPaths []string,
	inputArtifactTypes []artifact.Type,
	outputArtifactTypes []artifact.Type,
	storageConfig *shared.StorageConfig,
	jobManager job.JobManager,
	blacklistedOutputs []string,
) (string, error) {
	entryPoint := fn.EntryPoint
	if entryPoint == nil {
		entryPoint = &function.EntryPoint{
			File:      defaultFunctionEntryPointFile,
			ClassName: defaultFunctionEntryPointClass,
			Method:    defaultFunctionEntryPointMethod,
		}
	}

	jobName := generateFunctionJobName()

	jobSpec := job.NewFunctionSpec(
		jobName,
		storageConfig,
		metadataPath,
		fn.StoragePath,
		entryPoint.File,
		entryPoint.ClassName,
		entryPoint.Method,
		fn.CustomArgs,
		inputContentPaths,
		inputMetadataPaths,
		outputContentPaths,
		outputMetadataPaths,
		inputArtifactTypes,
		outputArtifactTypes,
		blacklistedOutputs,
	)

	err := jobManager.Launch(ctx, jobName, jobSpec)
	if err != nil {
		return "", errors.Wrap(err, "Unable to schedule function.")
	}

	return jobName, nil
}
