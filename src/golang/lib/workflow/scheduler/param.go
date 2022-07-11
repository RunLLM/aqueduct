package scheduler

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/operator/param"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

func generateParamJobName() string {
	return fmt.Sprintf("param-operator-%s", uuid.New().String())
}

func ScheduleParam(
	ctx context.Context,
	spec param.Param,
	metadataPath string,
	outputContentPath string,
	outputMetadataPath string,
	storageConfig *shared.StorageConfig,
	jobManager job.JobManager,
) (string, error) {
	jobName := generateParamJobName()

	jobSpec := job.NewParamSpec(
		jobName,
		storageConfig,
		metadataPath,
		spec.Val,
		outputContentPath,
		outputMetadataPath,
	)
	err := jobManager.Launch(ctx, jobName, jobSpec)
	if err != nil {
		return "", errors.Wrap(err, "Unable to schedule function.")
	}

	return jobName, nil
}
