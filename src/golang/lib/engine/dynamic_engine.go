package engine

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
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

func CreateDynamicEngine(
	ctx context.Context,
	engineIntegration *models.Integration,
	integrationRepo repos.Integration,
	vaultObject vault.Vault,
	db database.Database,
) error {
	dir := "/home/ubuntu/terraform/cgwu/learn-terraform-provision-eks-cluster"

	engineIntegration.Config["status"] = string(shared.K8sClusterCreatingStatus)
	_, err := integrationRepo.Update(
		ctx,
		engineIntegration.ID,
		map[string]interface{}{
			models.IntegrationConfig: &(engineIntegration.Config),
		},
		db,
	)
	if err != nil {
		return errors.Wrapf(err, "Failed to update Kubernetes cluster status to %s", engineIntegration.Config["status"])
	}

	log.Info("Running Terraform init...")
	if err := RunCmd("terraform", []string{"init"}, dir); err != nil {
		return errors.Wrap(err, "Terraform init failed")
	}

	// Fetch AWS credentials.
	if _, ok := engineIntegration.Config["cloud_integration_id"]; !ok {
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

	log.Info("Running Terraform apply...")
	if err := RunCmd(
		"terraform",
		[]string{"apply", "-auto-approve", accessKeyVar, secretAccessKeyVar},
		dir,
	); err != nil {
		return errors.Wrap(err, "Terraform apply failed")
	}

	if err := RunCmd(
		"aws",
		[]string{
			"eks",
			"update-kubeconfig",
			"--region",
			job.DefaultAwsRegion,
			"--name",
			engineIntegration.Config["cluster_name"],
		},
		dir,
	); err != nil {
		return errors.Wrap(err, "Failed to update Kubeconfig")
	}

	engineIntegration.Config["status"] = string(shared.K8sClusterActiveStatus)
	_, err = integrationRepo.Update(
		ctx,
		engineIntegration.ID,
		map[string]interface{}{
			models.IntegrationConfig: &(engineIntegration.Config),
		},
		db,
	)
	if err != nil {
		return errors.Wrapf(err, "Failed to update Kubernetes cluster status to %s", engineIntegration.Config["status"])
	}

	return nil
}

func DeleteDynamicEngine(
	ctx context.Context,
	engineIntegration *models.Integration,
	integrationRepo repos.Integration,
	db database.Database,
) error {
	engineIntegration.Config["status"] = string(shared.K8sClusterTerminatingStatus)
	if _, err := integrationRepo.Update(
		ctx,
		engineIntegration.ID,
		map[string]interface{}{
			models.IntegrationConfig: &(engineIntegration.Config),
		},
		db,
	); err != nil {
		return errors.Wrapf(err, "Failed to update Kubernetes cluster status to %s", engineIntegration.Config["status"])
	}

	dir := "/home/ubuntu/terraform/cgwu/learn-terraform-provision-eks-cluster"
	if err := RunCmd(
		"terraform",
		[]string{
			"destroy",
			"-auto-approve",
			"-var=access_key=",
			"-var=secret_key=",
		},
		dir,
	); err != nil {
		return errors.Wrap(err, "Unable to destroy k8s cluster")
	}

	if err := RunCmd("rm", []string{engineIntegration.Config["kubeconfig_path"]}, "."); err != nil {
		return errors.Wrap(err, "Unable to delete kubeconfig file")
	}

	engineIntegration.Config["status"] = string(shared.K8sClusterTerminatedStatus)
	if _, err := integrationRepo.Update(
		ctx,
		engineIntegration.ID,
		map[string]interface{}{
			models.IntegrationConfig: &(engineIntegration.Config),
		},
		db,
	); err != nil {
		return errors.Wrapf(err, "Failed to update Kubernetes cluster status to %s", engineIntegration.Config["status"])
	}

	return nil
}

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
	engineIntegration.Config["last_used_timestamp"] = strconv.FormatInt(currTimestamp, 10)
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

func RunCmd(command string, args []string, dir string) error {
	// create a new command
	cmd := exec.Command(command, args...)
	cmd.Dir = dir

	// create pipes for the command's standard output and standard error
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "Error creating stdout pipe")
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return errors.Wrap(err, "Error creating stderr pipe")
	}

	// start the command
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "Error starting command")
	}

	// create scanners to read from the pipes
	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)

	// start separate goroutines to stream the output from each scanner
	go func() {
		for stdoutScanner.Scan() {
			log.Infof("stdout: %s", stdoutScanner.Text())
		}
	}()
	go func() {
		for stderrScanner.Scan() {
			log.Errorf("stderr: %s", stderrScanner.Text())
		}
	}()

	// wait for the command to complete
	if err := cmd.Wait(); err != nil {
		return errors.Wrap(err, "Error waiting for command to complete")
	}

	return nil
}
