package dynamic

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var terraformDir = filepath.Join(os.Getenv("HOME"), ".aqueduct", "server", "dynamic", "terraform")

func PrepareEngine(
	ctx context.Context,
	engineIntegrationId uuid.UUID,
	integrationRepo repos.Integration,
	vaultObject vault.Vault,
	db database.Database,
) error {
	engineIntegration, err := UpdateEngineLastUsedTimestamp(
		ctx,
		engineIntegrationId,
		integrationRepo,
		db,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to update engine last used timestamp")
	}

	for {
		if engineIntegration.Config[shared.K8sStatusKey] == string(shared.K8sClusterTerminatedStatus) {
			log.Info("Kubernetes cluster is currently terminated, starting...")
			return CreateDynamicEngine(
				ctx,
				engineIntegration,
				integrationRepo,
				vaultObject,
				db,
			)
		} else if engineIntegration.Config[shared.K8sStatusKey] == string(shared.K8sClusterActiveStatus) {
			log.Info("Kubernetes cluster is currently active, proceeding...")
			return nil
		} else {
			log.Infof("Kubernetes cluster is currently in %s status. Waiting for 30 seconds before checking again...", engineIntegration.Config["status"])
			time.Sleep(30 * time.Second)

			engineIntegration, err = integrationRepo.Get(
				ctx,
				engineIntegrationId,
				db,
			)
			if err != nil {
				return errors.Wrap(err, "Failed to retrieve engine integration")
			}
		}
	}
}

// CreateDynamicEngine does the following:
// 1. Update the dynamic integration's DB record: set config["status"] to "Creating".
// 2. Run Terraform to create the cluster.
// 3. Update the kubeconfig file.
// 4. Update the dynamic integration's DB record: set config["status"] to "Active".
// If any step fails, it returns an error.
func CreateDynamicEngine(
	ctx context.Context,
	engineIntegration *models.Integration,
	integrationRepo repos.Integration,
	vaultObject vault.Vault,
	db database.Database,
) error {
	engineIntegration.Config[shared.K8sStatusKey] = string(shared.K8sClusterCreatingStatus)
	_, err := integrationRepo.Update(
		ctx,
		engineIntegration.ID,
		map[string]interface{}{
			models.IntegrationConfig: &(engineIntegration.Config),
		},
		db,
	)
	if err != nil {
		return errors.Wrapf(err, "Failed to update Kubernetes cluster status to %s", engineIntegration.Config[shared.K8sStatusKey])
	}

	if _, _, err := lib_utils.RunCmd("terraform", []string{"init"}, terraformDir, true); err != nil {
		return errors.Wrap(err, "Terraform init failed")
	}

	// Fetch AWS credentials.
	if _, ok := engineIntegration.Config[shared.K8sCloudIntegrationIdKey]; !ok {
		return errors.New("No cloud integration ID found in the engine integration object.")
	}
	cloudIntegrationId, err := uuid.Parse(engineIntegration.Config["cloud_integration_id"])
	if err != nil {
		return errors.Wrap(err, "Failed to parse cloud integration ID")
	}

	config, err := auth.ReadConfigFromSecret(ctx, cloudIntegrationId, vaultObject)
	if err != nil {
		return errors.Wrap(err, "Unable to read cloud integration config from vault.")
	}

	awsConfig, err := lib_utils.ParseAWSConfig(config)
	if err != nil {
		return errors.Wrap(err, "Unable to parse AWS config.")
	}

	accessKeyVar := fmt.Sprintf("-var=access_key=%s", awsConfig.AccessKeyId)
	secretAccessKeyVar := fmt.Sprintf("-var=secret_key=%s", awsConfig.SecretAccessKey)

	if _, _, err := lib_utils.RunCmd(
		"terraform",
		[]string{"apply", "-auto-approve", accessKeyVar, secretAccessKeyVar},
		terraformDir,
		true,
	); err != nil {
		return errors.Wrap(err, "Terraform apply failed")
	}

	if _, _, err := lib_utils.RunCmd(
		"aws",
		[]string{
			"eks",
			"update-kubeconfig",
			"--region",
			job.DefaultAwsRegion,
			"--name",
			engineIntegration.Config[shared.K8sClusterNameKey],
			"--kubeconfig",
			engineIntegration.Config[shared.K8sKubeconfigPathKey],
		},
		terraformDir,
		true,
	); err != nil {
		return errors.Wrap(err, "Failed to update Kubeconfig")
	}

	engineIntegration.Config[shared.K8sStatusKey] = string(shared.K8sClusterActiveStatus)
	_, err = integrationRepo.Update(
		ctx,
		engineIntegration.ID,
		map[string]interface{}{
			models.IntegrationConfig: &(engineIntegration.Config),
		},
		db,
	)
	if err != nil {
		return errors.Wrapf(err, "Failed to update Kubernetes cluster status to %s", engineIntegration.Config[shared.K8sStatusKey])
	}

	return nil
}

// DeleteDynamicEngine does the following:
// 1. Update the dynamic integration's DB record: set config["status"] to "Terminating".
// 2. Run Terraform to delete the cluster.
// 3. Remove the kubeconfig file.
// 4. Update the dynamic integration's DB record: set config["status"] to "Terminated".
// If any step fails, it returns an error.
func DeleteDynamicEngine(
	ctx context.Context,
	engineIntegration *models.Integration,
	integrationRepo repos.Integration,
	db database.Database,
) error {
	engineIntegration.Config[shared.K8sStatusKey] = string(shared.K8sClusterTerminatingStatus)
	if _, err := integrationRepo.Update(
		ctx,
		engineIntegration.ID,
		map[string]interface{}{
			models.IntegrationConfig: &(engineIntegration.Config),
		},
		db,
	); err != nil {
		return errors.Wrapf(err, "Failed to update Kubernetes cluster status to %s", engineIntegration.Config[shared.K8sStatusKey])
	}

	if _, _, err := lib_utils.RunCmd(
		"terraform",
		[]string{
			"destroy",
			"-auto-approve",
			"-var=access_key=",
			"-var=secret_key=",
		},
		terraformDir,
		true,
	); err != nil {
		return errors.Wrap(err, "Unable to destroy k8s cluster")
	}

	if _, _, err := lib_utils.RunCmd(
		"rm",
		[]string{engineIntegration.Config[shared.K8sKubeconfigPathKey]},
		".",
		true,
	); err != nil {
		return errors.Wrap(err, "Unable to delete kubeconfig file")
	}

	engineIntegration.Config[shared.K8sStatusKey] = string(shared.K8sClusterTerminatedStatus)
	if _, err := integrationRepo.Update(
		ctx,
		engineIntegration.ID,
		map[string]interface{}{
			models.IntegrationConfig: &(engineIntegration.Config),
		},
		db,
	); err != nil {
		return errors.Wrapf(err, "Failed to update Kubernetes cluster status to %s", engineIntegration.Config[shared.K8sStatusKey])
	}

	return nil
}

// UpdateEngineLastUsedTimestamp updates the dynamic integration's DB record:
// set config["last_used_timestamp"] to the current timestamp.
func UpdateEngineLastUsedTimestamp(
	ctx context.Context,
	engineIntegrationId uuid.UUID,
	integrationRepo repos.Integration,
	db database.Database,
) (*models.Integration, error) {
	engineIntegration, err := integrationRepo.Get(
		ctx,
		engineIntegrationId,
		db,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to retrieve engine integration")
	}

	currTimestamp := time.Now().Unix()
	engineIntegration.Config[shared.K8sLastUsedTimestampKey] = strconv.FormatInt(currTimestamp, 10)
	_, err = integrationRepo.Update(
		ctx,
		engineIntegration.ID,
		map[string]interface{}{
			models.IntegrationConfig: &(engineIntegration.Config),
		},
		db,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to update Kubernetes cluster's last used timestamp")
	}

	return engineIntegration, nil
}
