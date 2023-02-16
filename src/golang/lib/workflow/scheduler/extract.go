package scheduler

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator/connector"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

func generateExtractJobName() string {
	return fmt.Sprintf("extract-operator-%s", uuid.New().String())
}

func ScheduleExtract(
	ctx context.Context,
	spec connector.Extract,
	metadataPath string,
	inputParamNames []string,
	inputContentPaths []string,
	inputMetadataPaths []string,
	outputContentPath string,
	outputMetadataPath string,
	storageConfig *shared.StorageConfig,
	jobManager job.JobManager,
	vaultObject vault.Vault,
) (string, error) {
	config, err := auth.ReadConfigFromSecret(ctx, spec.IntegrationId, vaultObject)
	if err != nil {
		return "", err
	}

	jobName := generateExtractJobName()
	jobSpec := job.NewExtractSpec(
		jobName,
		storageConfig,
		metadataPath,
		spec.Service,
		config,
		spec.Parameters,
		inputParamNames,
		inputContentPaths,
		inputMetadataPaths,
		outputContentPath,
		outputMetadataPath,
	)

	err = jobManager.Launch(ctx, jobName, jobSpec)
	if err != nil {
		return "", errors.Wrap(err, "Unable to schedule Extract.")
	}

	return jobName, nil
}
