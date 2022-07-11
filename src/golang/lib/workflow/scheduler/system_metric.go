package scheduler

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/operator/system_metric"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

func generateSystemMetricJobName() string {
	return fmt.Sprintf("system_metric-operator-%s", uuid.New().String())
}

func ScheduleSystemMetric(
	ctx context.Context,
	spec system_metric.SystemMetric,
	metadataPath string,
	inputMetadataPaths []string,
	outputContentPath string,
	outputMetadataPath string,
	storageConfig *shared.StorageConfig,
	jobManager job.JobManager,
) (string, error) {
	jobName := generateSystemMetricJobName()

	jobSpec := job.NewSystemMetricSpec(
		jobName,
		storageConfig,
		metadataPath,
		spec.MetricName,
		inputMetadataPaths,
		outputContentPath,
		outputMetadataPath,
	)
	err := jobManager.Launch(ctx, jobName, jobSpec)
	if err != nil {
		return "", errors.Wrap(err, "Unable to schedule function.")
	}

	return jobName, nil
}
