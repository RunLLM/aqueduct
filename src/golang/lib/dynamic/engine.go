package dynamic

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/k8s"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type k8sClusterActionType string

const (
	k8sClusterCreateAction k8sClusterActionType = "create"
	k8sClusterUpdateAction k8sClusterActionType = "update"
)

const stateLockErrMsg = "Error acquiring the state lock"

var TerraformDir = filepath.Join(os.Getenv("HOME"), ".aqueduct", "server", "dynamic", "terraform")

// PrepareEngine
func PrepareEngine(
	ctx context.Context,
	configDelta map[string]string,
	engineIntegrationId uuid.UUID,
	integrationRepo repos.Integration,
	vaultObject vault.Vault,
	db database.Database,
) error {
	engineIntegration, err := integrationRepo.Get(
		ctx,
		engineIntegrationId,
		db,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to retrieve engine integration")
	}

	for {
		if engineIntegration.Config[shared.K8sStatusKey] == string(shared.K8sClusterTerminatedStatus) {
			log.Info("Kubernetes cluster is currently terminated, starting...")
			return CreateOrUpdateK8sCluster(
				ctx,
				configDelta,
				k8sClusterCreateAction,
				engineIntegration,
				integrationRepo,
				vaultObject,
				db,
			)
		} else if engineIntegration.Config[shared.K8sStatusKey] == string(shared.K8sClusterActiveStatus) {
			if len(configDelta) == 0 {
				log.Info("Kubernetes cluster is currently active, proceeding...")
				return nil
			} else {
				log.Info("Kubernetes cluster is currently active, updating the cluster since a non-empty config delta is provided...")
				return CreateOrUpdateK8sCluster(
					ctx,
					configDelta,
					k8sClusterUpdateAction,
					engineIntegration,
					integrationRepo,
					vaultObject,
					db,
				)
			}
		} else {
			if err := ResyncClusterState(ctx, engineIntegration, integrationRepo, db); err != nil {
				return errors.Wrap(err, "Failed to resync cluster state")
			}

			if engineIntegration.Config[shared.K8sStatusKey] == string(shared.K8sClusterTerminatedStatus) {
				// This means the cluster state is resynced to Terminated, so no need to wait 30 seconds.
				continue
			}

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

// CreateOrUpdateK8sCluster does the following:
//  1. If configDelta is not empty, apply the delta to engineIntegration.Config.
//  2. Update the dynamic integration's DB record: set config["status"] to "Creating" or "Updating".
//  3. Run terraform apply to create the cluster.
//  4. Update the kubeconfig file (only for "create" action).
//  5. Update the dynamic integration's DB record: set config["status"] to "Active", update
//     config["last_used_timestamp"] and update config to include the configDelta.
//
// If any step fails, it returns an error.
func CreateOrUpdateK8sCluster(
	ctx context.Context,
	configDelta map[string]string,
	action k8sClusterActionType, // can either be k8sClusterCreateAction or k8sClusterUpdateAction
	engineIntegration *models.Integration,
	integrationRepo repos.Integration,
	vaultObject vault.Vault,
	db database.Database,
) error {
	if !(action == k8sClusterCreateAction || action == k8sClusterUpdateAction) {
		return errors.Newf("Unsupport action %s.", action)
	}

	if action == k8sClusterUpdateAction && len(configDelta) == 0 {
		return nil // if there is no config delta, we don't need to update anything
	}

	if len(configDelta) > 0 {
		// Update config to include the new values.
		for key, value := range configDelta {
			if _, ok := engineIntegration.Config[key]; ok {
				engineIntegration.Config[key] = value
			} else {
				return errors.Newf("Config key %s is not supported.", key)
			}
		}

		if err := checkIfValidConfig(engineIntegration.Config); err != nil {
			return err
		}
	}

	var clusterStatus shared.K8sClusterStatusType
	if action == k8sClusterCreateAction {
		clusterStatus = shared.K8sClusterCreatingStatus
	} else {
		clusterStatus = shared.K8sClusterUpdatingStatus
	}

	if err := updateClusterStatus(ctx, clusterStatus, engineIntegration.ID, integrationRepo, db); err != nil {
		return err
	}

	if err := runTerraformApply(ctx, engineIntegration, vaultObject); err != nil {
		return err
	}

	if action == k8sClusterCreateAction {
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
			TerraformDir,
			true,
		); err != nil {
			return errors.Wrap(err, "Failed to update Kubeconfig")
		}
	}

	// We initialize the last used timestamp after the creation succeeded.
	if err := UpdateClusterLastUsedTimestamp(
		ctx,
		engineIntegration.ID,
		integrationRepo,
		db,
	); err != nil {
		return err
	}

	if err := updateClusterStatus(ctx, shared.K8sClusterActiveStatus, engineIntegration.ID, integrationRepo, db); err != nil {
		return err
	}

	// Finally, we update the database record to reflect the new config.
	if err := updateClusterConfig(ctx, configDelta, engineIntegration.ID, integrationRepo, db); err != nil {
		return err
	}

	return nil
}

// DeleteK8sCluster does the following:
// 1. Update the dynamic integration's DB record: set config["status"] to "Terminating".
// 2. Run Terraform to delete the cluster.
// 3. Remove the kubeconfig file.
// 4. Update the dynamic integration's DB record: set config["status"] to "Terminated".
// If any step fails, it returns an error.
func DeleteK8sCluster(
	ctx context.Context,
	force bool,
	engineIntegration *models.Integration,
	integrationRepo repos.Integration,
	db database.Database,
) error {
	if !force {
		safe, err := safeToDeleteCluster(ctx, engineIntegration)
		if err != nil {
			log.Errorf("We ran into an unexpected error: %v. Since the cluster might be in a bad state, we are force deleting it to be safe.", err)
		} else {
			if !safe {
				return errors.New("The k8s cluster cannot be deleted because there are pods still running.")
			}
		}
	}

	if err := updateClusterStatus(ctx, shared.K8sClusterTerminatingStatus, engineIntegration.ID, integrationRepo, db); err != nil {
		return err
	}

	// For deletion, we just need to initialize all variables to empty values.
	terraformArgs := generateTerraformVariables(&shared.AWSConfig{}, map[string]string{})

	if _, _, err := lib_utils.RunCmd(
		"terraform",
		append([]string{"destroy", "-auto-approve"}, terraformArgs...),
		TerraformDir,
		true,
	); err != nil {
		return errors.Wrap(err, "Unable to destroy k8s cluster")
	}

	kubeconfigFile := engineIntegration.Config[shared.K8sKubeconfigPathKey]
	if _, err := os.Stat(kubeconfigFile); !os.IsNotExist(err) {
		if _, _, err := lib_utils.RunCmd(
			"rm",
			[]string{kubeconfigFile},
			".",
			true,
		); err != nil {
			return errors.Wrap(err, "Unable to delete kubeconfig file")
		}
	}

	if err := updateClusterStatus(ctx, shared.K8sClusterTerminatedStatus, engineIntegration.ID, integrationRepo, db); err != nil {
		return err
	}

	return nil
}

// UpdateClusterLastUsedTimestamp updates the dynamic integration's DB record:
// set config["last_used_timestamp"] to the current timestamp.
func UpdateClusterLastUsedTimestamp(
	ctx context.Context,
	engineIntegrationId uuid.UUID,
	integrationRepo repos.Integration,
	db database.Database,
) error {
	engineIntegration, err := integrationRepo.Get(
		ctx,
		engineIntegrationId,
		db,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to retrieve engine integration")
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
		return errors.Wrap(err, "Failed to update Kubernetes cluster's last used timestamp")
	}

	return nil
}

// updateClusterStatus updates the dynamic integration's DB record:
// set config["status"] to the specified status.
func updateClusterStatus(
	ctx context.Context,
	status shared.K8sClusterStatusType,
	engineIntegrationId uuid.UUID,
	integrationRepo repos.Integration,
	db database.Database,
) error {
	engineIntegration, err := integrationRepo.Get(
		ctx,
		engineIntegrationId,
		db,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to retrieve engine integration")
	}

	engineIntegration.Config[shared.K8sStatusKey] = string(status)
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

// updateClusterConfig updates the dynamic integration's DB record:
// set config according to the config delta.
func updateClusterConfig(
	ctx context.Context,
	configDelta map[string]string,
	engineIntegrationId uuid.UUID,
	integrationRepo repos.Integration,
	db database.Database,
) error {
	if len(configDelta) == 0 {
		return nil
	}

	engineIntegration, err := integrationRepo.Get(
		ctx,
		engineIntegrationId,
		db,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to retrieve engine integration")
	}

	// Update config to include the new values.
	for key, value := range configDelta {
		if _, ok := engineIntegration.Config[key]; ok {
			engineIntegration.Config[key] = value
		} else {
			return errors.Newf("Config key %s is not supported.", key)
		}
	}

	_, err = integrationRepo.Update(
		ctx,
		engineIntegration.ID,
		map[string]interface{}{
			models.IntegrationConfig: &(engineIntegration.Config),
		},
		db,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to update Kubernetes cluster config")
	}

	return nil
}

// ResyncClusterState does the following: when the database state of the k8s cluster is not
// Active or Terminating, it checks whether there is a terraform action happening. If not, this means
// we are in an inconsistent state due to server failure or race condition. If so, we resync the
// database state with terraform state by deleting the cluster and updating the database state to be
// Terminated.
func ResyncClusterState(
	ctx context.Context,
	engineIntegration *models.Integration,
	integrationRepo repos.Integration,
	db database.Database,
) error {
	if engineIntegration.Config[shared.K8sStatusKey] == string(shared.K8sClusterActiveStatus) || engineIntegration.Config[shared.K8sStatusKey] == string(shared.K8sClusterTerminatedStatus) {
		log.Infof("No need to resync state because the cluster status is %s", engineIntegration.Config[shared.K8sStatusKey])
		return nil
	}

	// Terraform does not offer an API to check if the state is locked, but we can use `terraform plan`
	// as a workaround: we know the state is locked if the stderr contains stateLockErrMsg.
	// If the state is locked, we know there's an ongoing action (terraform apply or terraform destroy)
	// happening, which is the expected case here and so we return with no error.
	if _, stderr, err := lib_utils.RunCmd(
		"terraform",
		[]string{
			"plan",
		},
		TerraformDir,
		false,
	); err != nil {
		if strings.Contains(stderr, stateLockErrMsg) {
			return nil
		}
	}

	// If we reach here, it means although the database state tells us there should be some terraform
	// action happening, there isn't. This can happen due to server failure, which creates an
	// inconsistent state between the database and terraform. In this case, we resync the state by
	// deleting the cluster and updating the database state to be Terminated.
	log.Error("Dynamic k8s cluster might be in an inconsistent state. Resolving state by deleting the cluster...")
	return DeleteK8sCluster(ctx, true, engineIntegration, integrationRepo, db)
}

// safeToDeleteCluster checks whether there are pods in the aqueduct namespace of the dynamic k8s
// cluster that are in ContainerCreating or Running status. If so, it returns false. Otherwise it
// returns true.
func safeToDeleteCluster(
	ctx context.Context,
	engineIntegration *models.Integration,
) (bool, error) {
	useSameCluster, err := strconv.ParseBool(engineIntegration.Config[shared.K8sUseSameClusterKey])
	if err != nil {
		return false, errors.Wrap(err, "Error parsing use_same_cluster flag")
	}

	k8sClient, err := k8s.CreateK8sClient(engineIntegration.Config[shared.K8sKubeconfigPathKey], useSameCluster)
	if err != nil {
		return false, errors.Wrap(err, "Error while creating K8sClient")
	}

	pods, err := k8sClient.CoreV1().Pods(k8s.AqueductNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, errors.Wrap(err, "Error while listing pods in the aqueduct namespace")
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning {
			return false, nil
		}

		if pod.Status.Phase == v1.PodPending {
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.State.Waiting != nil && cs.State.Waiting.Reason == "ContainerCreating" {
					return false, nil
				}
			}
		}
	}

	return true, nil
}

func generateTerraformVariables(
	awsConfig *shared.AWSConfig,
	engineConfig map[string]string,
) []string {
	accessKeyVar := fmt.Sprintf("-var=access_key=%s", awsConfig.AccessKeyId)
	secretAccessKeyVar := fmt.Sprintf("-var=secret_key=%s", awsConfig.SecretAccessKey)
	regionVar := fmt.Sprintf("-var=region=%s", awsConfig.Region)

	cpuNodeTypeVar := fmt.Sprintf("-var=cpu_node_type=%s", engineConfig[shared.K8sCpuNodeTypeKey])
	gpuNodeTypeVar := fmt.Sprintf("-var=gpu_node_type=%s", engineConfig[shared.K8sGpuNodeTypeKey])
	minCpuNodeVar := fmt.Sprintf("-var=min_cpu_node=%s", engineConfig[shared.K8sMinCpuNodeKey])
	maxCpuNodeVar := fmt.Sprintf("-var=max_cpu_node=%s", engineConfig[shared.K8sMaxCpuNodeKey])
	minGpuNodeVar := fmt.Sprintf("-var=min_gpu_node=%s", engineConfig[shared.K8sMinGpuNodeKey])
	maxGpuNodeVar := fmt.Sprintf("-var=max_gpu_node=%s", engineConfig[shared.K8sMaxGpuNodeKey])

	return []string{
		accessKeyVar,
		secretAccessKeyVar,
		regionVar,
		cpuNodeTypeVar,
		gpuNodeTypeVar,
		minCpuNodeVar,
		maxCpuNodeVar,
		minGpuNodeVar,
		maxGpuNodeVar,
	}
}

func runTerraformApply(
	ctx context.Context,
	engineIntegration *models.Integration,
	vaultObject vault.Vault,
) error {
	// Fetch AWS credentials.
	if _, ok := engineIntegration.Config[shared.K8sCloudIntegrationIdKey]; !ok {
		return errors.New("No cloud integration ID found in the engine integration object.")
	}
	cloudIntegrationId, err := uuid.Parse(engineIntegration.Config[shared.K8sCloudIntegrationIdKey])
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

	terraformArgs := generateTerraformVariables(awsConfig, engineIntegration.Config)

	if _, _, err := lib_utils.RunCmd(
		"terraform",
		append([]string{"apply", "-auto-approve"}, terraformArgs...),
		TerraformDir,
		true,
	); err != nil {
		return errors.Wrap(err, "Terraform apply failed")
	}

	return nil
}

func checkIfValidConfig(config map[string]string) error {
	// We require a minimum keepalive period of 10min (600 seconds).
	keepalive, err := strconv.Atoi(config[shared.K8sKeepaliveKey])
	if err != nil {
		return errors.Wrap(err, "Error parsing keepalive value")
	}

	if keepalive < shared.K8sMinimumKeepalive {
		return errors.Newf("The new keepalive value %d is smaller than the minimum value %d", keepalive, shared.K8sMinimumKeepalive)
	}

	minCpuNode, err := strconv.Atoi(config[shared.K8sMinCpuNodeKey])
	if err != nil {
		return errors.Wrap(err, "Error parsing min CPU node value")
	}

	maxCpuNode, err := strconv.Atoi(config[shared.K8sMaxCpuNodeKey])
	if err != nil {
		return errors.Wrap(err, "Error parsing max CPU node value")
	}

	if maxCpuNode < minCpuNode {
		return errors.Newf("The new max CPU node value %d is smaller than the min CPU node value %d", maxCpuNode, minCpuNode)
	}

	minGpuNode, err := strconv.Atoi(config[shared.K8sMinGpuNodeKey])
	if err != nil {
		return errors.Wrap(err, "Error parsing min GPU node value")
	}

	maxGpuNode, err := strconv.Atoi(config[shared.K8sMaxGpuNodeKey])
	if err != nil {
		return errors.Wrap(err, "Error parsing max GPU node value")
	}

	if maxGpuNode < minGpuNode {
		return errors.Newf("The new max GPU node value %d is smaller than the min GPU node value %d", maxGpuNode, minGpuNode)
	}

	return nil
}
