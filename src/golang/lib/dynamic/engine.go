package dynamic

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aqueducthq/aqueduct/lib/database"
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
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type k8sClusterActionType string

const (
	K8sClusterCreateAction k8sClusterActionType = "create"
	K8sClusterUpdateAction k8sClusterActionType = "update"
)

const (
	stateLockErrMsg       = "Error acquiring the state lock"
	K8sResourceNameSuffix = "aqueduct_ondemand_k8s"
)

var (
	EKSTerraformTemplateDir = filepath.Join(os.Getenv("HOME"), ".aqueduct", "server", "template", "aws", "eks")
	GKETerraformTemplateDir = filepath.Join(os.Getenv("HOME"), ".aqueduct", "server", "template", "gke")
)

// PrepareCluster blocks until the cluster is in status "Active".
func PrepareCluster(
	ctx context.Context,
	configDelta *shared.DynamicK8sConfig,
	engineResourceId uuid.UUID,
	resourceRepo repos.Resource,
	vaultObject vault.Vault,
	db database.Database,
) error {
	engineResource, err := resourceRepo.Get(
		ctx,
		engineResourceId,
		db,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to retrieve engine resource")
	}

	for {
		if engineResource.Config[shared.K8sStatusKey] == string(shared.K8sClusterTerminatedStatus) {
			log.Info("Kubernetes cluster is currently terminated, starting...")
			return CreateOrUpdateK8sCluster(
				ctx,
				configDelta,
				K8sClusterCreateAction,
				engineResource,
				resourceRepo,
				vaultObject,
				db,
			)
		} else if engineResource.Config[shared.K8sStatusKey] == string(shared.K8sClusterActiveStatus) {
			if len(configDelta.ToMap()) == 0 {
				log.Info("Kubernetes cluster is currently active, proceeding...")
				return nil
			} else {
				log.Info("Kubernetes cluster is currently active, updating the cluster since a non-empty config delta is provided...")
				return CreateOrUpdateK8sCluster(
					ctx,
					configDelta,
					K8sClusterUpdateAction,
					engineResource,
					resourceRepo,
					vaultObject,
					db,
				)
			}
		} else {
			engineResource, err = PollClusterStatus(ctx, engineResource, resourceRepo, vaultObject, db)
			if err != nil {
				return err
			}
		}
	}
}

// CreateOrUpdateK8sCluster does the following:
//  1. If configDelta is not empty, apply the delta to engineResourceConfig.
//  2. Update the dynamic resource's DB record: set config["status"] to "Creating" or "Updating".
//  3. Run terraform apply to create the cluster.
//  4. Update the kubeconfig file (only for "create" action).
//  5. Update the dynamic resource's DB record: set config["status"] to "Active", update
//     config["last_used_timestamp"] and update config to include the configDelta.
//
// If any step fails, it returns an error.
func CreateOrUpdateK8sCluster(
	ctx context.Context,
	configDelta *shared.DynamicK8sConfig,
	action k8sClusterActionType, // can either be k8sClusterCreateAction or k8sClusterUpdateAction
	engineResource *models.Resource,
	resourceRepo repos.Resource,
	vaultObject vault.Vault,
	db database.Database,
) error {
	if !(action == K8sClusterCreateAction || action == K8sClusterUpdateAction) {
		return errors.Newf("Unsupported action %s.", action)
	}

	configDeltaMap := configDelta.ToMap()

	if action == K8sClusterUpdateAction && len(configDeltaMap) == 0 {
		return nil // if there is no config delta, we don't need to update anything
	}

	if len(configDeltaMap) > 0 {
		// Update config to reflect the new values.
		for key, value := range configDeltaMap {
			engineResource.Config[key] = value
		}
	}

	if err := CheckIfValidConfig(action, engineResource.Config); err != nil {
		return err
	}

	var clusterStatus shared.K8sClusterStatusType
	if action == K8sClusterCreateAction {
		clusterStatus = shared.K8sClusterCreatingStatus
	} else {
		clusterStatus = shared.K8sClusterUpdatingStatus
	}

	if err := updateClusterStatus(ctx, clusterStatus, engineResource.ID, resourceRepo, db); err != nil {
		return err
	}

	var awsConfig *shared.AWSConfig
	var gcpConfig *shared.GCPConfig
	var err error

	if engineResource.Config[shared.K8sCloudProviderKey] == string(shared.GCPProvider) {
		gcpConfig, err = fetchGCPCredential(ctx, engineResource, vaultObject)
		if err != nil {
			return err
		}
	} else {
		awsConfig, err = fetchAWSCredential(ctx, engineResource, vaultObject)
		if err != nil {
			return err
		}
	}

	if err := runTerraformApply(awsConfig, gcpConfig, engineResource); err != nil {
		return err
	}

	if action == K8sClusterCreateAction {
		if awsConfig != nil {
			var envVars []string
			if awsConfig.AccessKeyId != "" && awsConfig.SecretAccessKey != "" && awsConfig.Region != "" {
				// If we enter here, it means the authentication mode is access key.
				envVars = []string{
					fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", awsConfig.AccessKeyId),
					fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", awsConfig.SecretAccessKey),
					fmt.Sprintf("AWS_REGION=%s", awsConfig.Region),
				}
			} else {
				// If we enter here, it means the authentication mode is credential file.
				envVars = []string{
					fmt.Sprintf("AWS_SHARED_CREDENTIALS_FILE=%s", awsConfig.ConfigFilePath),
					fmt.Sprintf("AWS_PROFILE=%s", awsConfig.ConfigFileProfile),
				}
			}
			if _, _, err := lib_utils.RunCmd(
				"env",
				append(
					envVars,
					"aws",
					"eks",
					"update-kubeconfig",
					"--name",
					engineResource.Config[shared.K8sClusterNameKey],
					"--kubeconfig",
					engineResource.Config[shared.K8sKubeconfigPathKey],
				),
				engineResource.Config[shared.K8sTerraformPathKey],
				true,
			); err != nil {
				return errors.Wrap(err, "Failed to update Kubeconfig")
			}

			config, err := clientcmd.LoadFromFile(engineResource.Config[shared.K8sKubeconfigPathKey])
			if err != nil {
				return errors.Wrap(err, "Failed to load Kubeconfig")
			}

			for _, authInfo := range config.AuthInfos {
				if awsConfig.AccessKeyId != "" && awsConfig.SecretAccessKey != "" && awsConfig.Region != "" {
					authInfo.Exec.Env = append(authInfo.Exec.Env, api.ExecEnvVar{
						Name:  "AWS_ACCESS_KEY_ID",
						Value: awsConfig.AccessKeyId,
					})
					authInfo.Exec.Env = append(authInfo.Exec.Env, api.ExecEnvVar{
						Name:  "AWS_SECRET_ACCESS_KEY",
						Value: awsConfig.SecretAccessKey,
					})
					authInfo.Exec.Env = append(authInfo.Exec.Env, api.ExecEnvVar{
						Name:  "AWS_REGION",
						Value: awsConfig.Region,
					})
				} else {
					authInfo.Exec.Env = append(authInfo.Exec.Env, api.ExecEnvVar{
						Name:  "AWS_SHARED_CREDENTIALS_FILE",
						Value: awsConfig.ConfigFilePath,
					})
					authInfo.Exec.Env = append(authInfo.Exec.Env, api.ExecEnvVar{
						Name:  "AWS_PROFILE",
						Value: awsConfig.ConfigFileProfile,
					})
				}
			}

			err = clientcmd.WriteToFile(*config, engineResource.Config[shared.K8sKubeconfigPathKey])
			if err != nil {
				return errors.Wrap(err, "Failed to update Kubeconfig with environment variables")
			}
		} else {
			// GCP
			var key struct {
				ProjectID   string `json:"project_id"`
				ClientEmail string `json:"client_email"`
			}

			err := json.Unmarshal([]byte(gcpConfig.ServiceAccountKey), &key)
			if err != nil {
				return errors.Wrap(err, "Failed to parse project ID and client email from service account key")
			}

			// Write the service account key to a temporary file in the resource's Terraform directory.
			// This is necessary because the gcloud CLI requires a file path to the service account key.
			serviceAccountKeyPath := filepath.Join(engineResource.Config[shared.K8sTerraformPathKey], "service_account_key.json")
			err = os.WriteFile(serviceAccountKeyPath, []byte(gcpConfig.ServiceAccountKey), 0o644)
			if err != nil {
				return errors.Wrap(err, "Failed to write service account key to temporary file")
			}

			if _, _, err := lib_utils.RunCmd(
				"gcloud",
				[]string{
					"auth",
					"activate-service-account",
					"--key-file",
					serviceAccountKeyPath,
				},
				"",
				true,
			); err != nil {
				return errors.Wrap(err, "Failed to activate service account")
			}

			if _, _, err := lib_utils.RunCmd(
				"gcloud",
				[]string{
					"config",
					"set",
					"account",
					key.ClientEmail,
				},
				"",
				true,
			); err != nil {
				return errors.Wrap(err, "Failed to set account")
			}

			kubeconfigEnv := fmt.Sprintf("KUBECONFIG=%s", engineResource.Config[shared.K8sKubeconfigPathKey])
			if _, _, err := lib_utils.RunCmd(
				"env",
				[]string{
					kubeconfigEnv,
					"gcloud",
					"container",
					"clusters",
					"get-credentials",
					engineResource.Config[shared.K8sClusterNameKey],
					"--region",
					gcpConfig.Region,
					"--project",
					key.ProjectID,
				},
				"",
				true,
			); err != nil {
				return errors.Wrap(err, "Failed to update Kubeconfig")
			}

			// Delete the temporary service account key file.
			if err := os.Remove(serviceAccountKeyPath); err != nil {
				return errors.Wrap(err, "Failed to remove temporary service account key file")
			}

			log.Info("Successfully created ondemand GKE cluster, waiting 10 seconds before proceeding...")
			time.Sleep(10 * time.Second)
		}
	}

	// We initialize the last used timestamp after the creation succeeded.
	if err := UpdateClusterLastUsedTimestamp(
		ctx,
		engineResource.ID,
		resourceRepo,
		db,
	); err != nil {
		return err
	}

	if err := updateClusterStatus(ctx, shared.K8sClusterActiveStatus, engineResource.ID, resourceRepo, db); err != nil {
		return err
	}

	// Finally, we update the database record to reflect the new config.
	if err := updateClusterConfig(ctx, action, configDeltaMap, engineResource.ID, resourceRepo, db); err != nil {
		return err
	}

	return nil
}

// DeleteK8sCluster does the following:
// 1. Update the dynamic resource's DB record: set config["status"] to "Terminating".
// 2. Run Terraform to delete the cluster.
// 3. Remove the kubeconfig file.
// 4. Update the dynamic resource's DB record: set config["status"] to "Terminated".
// If any step fails, it returns an error.
// If skipPodsStatusCheck is set to false, it checks whether there are pods in Running or ContainerCreating
// status and if so, reject the deletion request.
func DeleteK8sCluster(
	ctx context.Context,
	skipPodsStatusCheck bool,
	engineResource *models.Resource,
	resourceRepo repos.Resource,
	vaultObject vault.Vault,
	db database.Database,
) error {
	if !skipPodsStatusCheck {
		useSameCluster, err := strconv.ParseBool(engineResource.Config[shared.K8sUseSameClusterKey])
		if err != nil {
			return errors.Wrap(err, "Error parsing use_same_cluster flag")
		}

		safe, err := k8s.SafeToDeleteCluster(ctx, useSameCluster, engineResource.Config[shared.K8sKubeconfigPathKey])
		if err != nil {
			return err
		}

		if !safe {
			return errors.New("The k8s cluster cannot be deleted because there are pods still running.")
		}
	}

	if err := updateClusterStatus(ctx, shared.K8sClusterTerminatingStatus, engineResource.ID, resourceRepo, db); err != nil {
		return err
	}

	// Even for deletion, we need to specify the AWS region, so we need to pass in the cloud provider
	// config instead of a dummy one to generateTerraformVariables.
	var awsConfig *shared.AWSConfig
	var gcpConfig *shared.GCPConfig
	var err error

	if engineResource.Config[shared.K8sCloudProviderKey] == string(shared.GCPProvider) {
		gcpConfig, err = fetchGCPCredential(ctx, engineResource, vaultObject)
		if err != nil {
			return err
		}
	} else {
		awsConfig, err = fetchAWSCredential(ctx, engineResource, vaultObject)
		if err != nil {
			return err
		}
	}

	terraformArgs, err := generateTerraformVariables(awsConfig, gcpConfig, engineResource.Config)
	if err != nil {
		return err
	}

	if _, _, err := lib_utils.RunCmd(
		"terraform",
		append([]string{"destroy", "-auto-approve"}, terraformArgs...),
		engineResource.Config[shared.K8sTerraformPathKey],
		true,
	); err != nil {
		return errors.Wrap(err, "Unable to destroy k8s cluster")
	}

	kubeconfigFile := engineResource.Config[shared.K8sKubeconfigPathKey]
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

	if err := updateClusterStatus(ctx, shared.K8sClusterTerminatedStatus, engineResource.ID, resourceRepo, db); err != nil {
		return err
	}

	return nil
}

// UpdateClusterLastUsedTimestamp updates the dynamic resource's DB record:
// set config["last_used_timestamp"] to the current timestamp.
func UpdateClusterLastUsedTimestamp(
	ctx context.Context,
	engineResourceID uuid.UUID,
	resourceRepo repos.Resource,
	db database.Database,
) error {
	engineResource, err := resourceRepo.Get(
		ctx,
		engineResourceID,
		db,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to retrieve engine resource")
	}

	currTimestamp := time.Now().Unix()
	engineResource.Config[shared.K8sLastUsedTimestampKey] = strconv.FormatInt(currTimestamp, 10)
	_, err = resourceRepo.Update(
		ctx,
		engineResource.ID,
		map[string]interface{}{
			models.ResourceConfig: &(engineResource.Config),
		},
		db,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to update Kubernetes cluster's last used timestamp")
	}

	return nil
}

// updateClusterStatus updates the dynamic resource's DB record:
// set config["status"] to the specified status.
func updateClusterStatus(
	ctx context.Context,
	status shared.K8sClusterStatusType,
	engineResourceID uuid.UUID,
	resourceRepo repos.Resource,
	db database.Database,
) error {
	engineResource, err := resourceRepo.Get(
		ctx,
		engineResourceID,
		db,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to retrieve engine resource")
	}

	engineResource.Config[shared.K8sStatusKey] = string(status)
	_, err = resourceRepo.Update(
		ctx,
		engineResource.ID,
		map[string]interface{}{
			models.ResourceConfig: &(engineResource.Config),
		},
		db,
	)
	if err != nil {
		return errors.Wrapf(err, "Failed to update Kubernetes cluster status to %s", engineResource.Config[shared.K8sStatusKey])
	}

	return nil
}

// updateClusterConfig updates the dynamic resource's DB record:
// set config according to the config delta.
func updateClusterConfig(
	ctx context.Context,
	action k8sClusterActionType,
	configDeltaMap map[string]string,
	engineResourceID uuid.UUID,
	resourceRepo repos.Resource,
	db database.Database,
) error {
	if len(configDeltaMap) == 0 {
		return nil
	}

	engineResource, err := resourceRepo.Get(
		ctx,
		engineResourceID,
		db,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to retrieve engine resource")
	}

	// Update config to include the new values.
	for key, value := range configDeltaMap {
		engineResource.Config[key] = value
	}

	if action == K8sClusterCreateAction {
		// If this is a request to create a new cluster, we need to refresh the desired node counts.
		engineResource.Config[shared.K8sDesiredCpuNodeKey] = engineResource.Config[shared.K8sMinCpuNodeKey]
		engineResource.Config[shared.K8sDesiredGpuNodeKey] = engineResource.Config[shared.K8sMinGpuNodeKey]
	}

	_, err = resourceRepo.Update(
		ctx,
		engineResource.ID,
		map[string]interface{}{
			models.ResourceConfig: &(engineResource.Config),
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
	engineResource *models.Resource,
	resourceRepo repos.Resource,
	vaultObject vault.Vault,
	db database.Database,
) error {
	if engineResource.Config[shared.K8sStatusKey] == string(shared.K8sClusterActiveStatus) || engineResource.Config[shared.K8sStatusKey] == string(shared.K8sClusterTerminatedStatus) {
		log.Infof("No need to resync state because the cluster status is %s", engineResource.Config[shared.K8sStatusKey])
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
		engineResource.Config[shared.K8sTerraformPathKey],
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
	return DeleteK8sCluster(
		ctx,
		true, // skipPodsStatusCheck
		engineResource,
		resourceRepo,
		vaultObject,
		db,
	)
}

func PollClusterStatus(
	ctx context.Context,
	engineResource *models.Resource,
	resourceRepo repos.Resource,
	vaultObject vault.Vault,
	db database.Database,
) (*models.Resource, error) {
	if err := ResyncClusterState(ctx, engineResource, resourceRepo, vaultObject, db); err != nil {
		return nil, errors.Wrap(err, "Failed to resync cluster state")
	}

	engineResource, err := resourceRepo.Get(
		ctx,
		engineResource.ID,
		db,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to retrieve engine resource")
	}

	if engineResource.Config[shared.K8sStatusKey] == string(shared.K8sClusterTerminatedStatus) {
		// This means the cluster state is resynced to Terminated, so no need to wait.
		return engineResource, nil
	}

	log.Infof("Kubernetes cluster is currently in %s status. Waiting for %d seconds before checking again...", engineResource.Config[shared.K8sStatusKey], shared.DynamicK8sClusterStatusPollPeriod)
	time.Sleep(shared.DynamicK8sClusterStatusPollPeriod * time.Second)

	engineResource, err = resourceRepo.Get(
		ctx,
		engineResource.ID,
		db,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to retrieve engine resource")
	}

	return engineResource, nil
}

func generateTerraformVariables(
	awsConfig *shared.AWSConfig,
	gcpConfig *shared.GCPConfig,
	engineConfig map[string]string,
) ([]string, error) {
	cpuNodeTypeVar := fmt.Sprintf("-var=cpu_node_type=%s", engineConfig[shared.K8sCpuNodeTypeKey])
	gpuNodeTypeVar := fmt.Sprintf("-var=gpu_node_type=%s", engineConfig[shared.K8sGpuNodeTypeKey])
	minCpuNodeVar := fmt.Sprintf("-var=min_cpu_node=%s", engineConfig[shared.K8sMinCpuNodeKey])
	maxCpuNodeVar := fmt.Sprintf("-var=max_cpu_node=%s", engineConfig[shared.K8sMaxCpuNodeKey])
	minGpuNodeVar := fmt.Sprintf("-var=min_gpu_node=%s", engineConfig[shared.K8sMinGpuNodeKey])
	maxGpuNodeVar := fmt.Sprintf("-var=max_gpu_node=%s", engineConfig[shared.K8sMaxGpuNodeKey])

	clusterNameVar := fmt.Sprintf("-var=cluster_name=%s", engineConfig[shared.K8sClusterNameKey])

	vars := []string{
		cpuNodeTypeVar,
		gpuNodeTypeVar,
		minCpuNodeVar,
		maxCpuNodeVar,
		minGpuNodeVar,
		maxGpuNodeVar,
		clusterNameVar,
	}

	if awsConfig != nil {
		accessKeyVar := fmt.Sprintf("-var=access_key=%s", awsConfig.AccessKeyId)
		secretAccessKeyVar := fmt.Sprintf("-var=secret_key=%s", awsConfig.SecretAccessKey)
		regionVar := fmt.Sprintf("-var=region=%s", awsConfig.Region)
		credentialPathVar := fmt.Sprintf("-var=credentials_file=%s", awsConfig.ConfigFilePath)
		profileVar := fmt.Sprintf("-var=profile=%s", awsConfig.ConfigFileProfile)

		if awsConfig.ConfigFilePath != "" && awsConfig.ConfigFileProfile != "" {
			// If the authentication mode is credential file, we need to retrieve the AWS region via
			// `aws configure get region` and explicitly pass it to Terraform.
			region, stderr, err := lib_utils.RunCmd(
				"env",
				[]string{
					fmt.Sprintf("AWS_SHARED_CREDENTIALS_FILE=%s", awsConfig.ConfigFilePath),
					fmt.Sprintf("AWS_PROFILE=%s", awsConfig.ConfigFileProfile),
					"aws",
					"configure",
					"get",
					"region",
				},
				"",
				false,
			)
			// We need to check if stderr is empty because when the region is not specified in the
			// profile, the cmd will error and it will produce an empty stdout and stderr. In this case,
			// we should just set the region to an empty string, which means using the default region.
			if err != nil && stderr != "" {
				return nil, err
			}

			regionVar = fmt.Sprintf("-var=region=%s", strings.TrimRight(region, "\n"))
		}

		vars = append(vars, []string{
			accessKeyVar,
			secretAccessKeyVar,
			regionVar,
			credentialPathVar,
			profileVar,
		}...)
	} else {
		var key struct {
			ProjectID string `json:"project_id"`
		}

		err := json.Unmarshal([]byte(gcpConfig.ServiceAccountKey), &key)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to parse project ID and client email from service account key")
		}

		regionVar := fmt.Sprintf("-var=region=%s", gcpConfig.Region)
		zoneVar := fmt.Sprintf("-var=zone=%s", gcpConfig.Zone)
		secretKeyVar := fmt.Sprintf("-var=secret_key=%s", gcpConfig.ServiceAccountKey)
		projectIDVar := fmt.Sprintf("-var=project_id=%s", key.ProjectID)

		vars = append(vars, []string{
			regionVar,
			zoneVar,
			secretKeyVar,
			projectIDVar,
		}...)
	}

	return vars, nil
}

func runTerraformApply(
	awsConfig *shared.AWSConfig,
	gcpConfig *shared.GCPConfig,
	engineResource *models.Resource,
) error {
	terraformArgs, err := generateTerraformVariables(awsConfig, gcpConfig, engineResource.Config)
	if err != nil {
		return err
	}

	if _, _, err := lib_utils.RunCmd(
		"terraform",
		append([]string{"apply", "-auto-approve"}, terraformArgs...),
		engineResource.Config[shared.K8sTerraformPathKey],
		true,
	); err != nil {
		errMsg := "Terraform apply failed. Note that if the error has to do with insufficient " +
			"permissions, you will need to add the missing permissions to your registered cloud account."
		return errors.Wrap(err, errMsg)
	}

	return nil
}

func CheckIfValidConfig(action k8sClusterActionType, config map[string]string) error {
	// We require a minimum keepalive period of 10 min (600 seconds).
	keepalive, err := strconv.Atoi(config[shared.K8sKeepaliveKey])
	if err != nil {
		return errors.Wrap(err, "Error parsing keepalive value")
	}

	if keepalive < shared.K8sMinimumKeepalive {
		return errors.Newf("The new keepalive value %d is smaller than the minimum allowed value %d", keepalive, shared.K8sMinimumKeepalive)
	}

	minCpuNode, err := strconv.Atoi(config[shared.K8sMinCpuNodeKey])
	if err != nil {
		return errors.Wrap(err, "Error parsing min CPU node value")
	}

	maxCpuNode, err := strconv.Atoi(config[shared.K8sMaxCpuNodeKey])
	if err != nil {
		return errors.Wrap(err, "Error parsing max CPU node value")
	}

	if maxCpuNode < 1 {
		return errors.Newf("Max CPU node value should be at least 1, got %d", maxCpuNode)
	}

	if minCpuNode < 0 {
		return errors.Newf("Min CPU node value should be at least 0, got %d", minCpuNode)
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

	if maxGpuNode < 1 {
		return errors.Newf("Max GPU node value should be at least 1, got %d", maxGpuNode)
	}

	if minGpuNode < 0 {
		return errors.Newf("Min GPU node value should be at least 0, got %d", minGpuNode)
	}

	if maxGpuNode < minGpuNode {
		return errors.Newf("The new max GPU node value %d is smaller than the min GPU node value %d", maxGpuNode, minGpuNode)
	}

	if action == K8sClusterUpdateAction {
		// We only check the constraint below for update, because for create, we are overwriting the desired node values.
		desiredCpuNode, err := strconv.Atoi(config[shared.K8sDesiredCpuNodeKey])
		if err != nil {
			return errors.Wrap(err, "Error parsing desired CPU node value")
		}

		if minCpuNode > desiredCpuNode {
			return errors.Newf("The new min CPU node value %d is bigger than the desired CPU node value %d. To increase the min value, please delete the cluster and re-create it with the new config", minCpuNode, desiredCpuNode)
		}

		if maxCpuNode < desiredCpuNode {
			return errors.Newf("The new max CPU node value %d is smaller than the desired CPU node value %d. To reduce the max value, please delete the cluster and re-create it with the new config", maxCpuNode, desiredCpuNode)
		}

		desiredGpuNode, err := strconv.Atoi(config[shared.K8sDesiredGpuNodeKey])
		if err != nil {
			return errors.Wrap(err, "Error parsing desired GPU node value")
		}

		if minGpuNode > desiredGpuNode {
			return errors.Newf("The new min GPU node value %d is bigger than the desired GPU node value %d. To increase the min value, please delete the cluster and re-create it with the new config", minGpuNode, desiredGpuNode)
		}

		if maxGpuNode < desiredGpuNode {
			return errors.Newf("The new max GPU node value %d is smaller than the desired GPU node value %d. To reduce the max value, please delete the cluster and re-create it with the new config", maxGpuNode, desiredGpuNode)
		}
	}

	return nil
}

// GenerateClusterName generates a EKS cluster name by concatenating aqueduct with a
// random string of length 16.
func GenerateClusterName() (string, error) {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, 16)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letterBytes))))
		if err != nil {
			return "", err
		}
		b[i] = letterBytes[n.Int64()]
	}

	return fmt.Sprintf("%s_%s", "aqueduct", string(b)), nil
}

func GenerateClusterNameGKE() (string, error) {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"

	b := make([]byte, 16)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letterBytes))))
		if err != nil {
			return "", err
		}
		b[i] = letterBytes[n.Int64()]
	}

	return fmt.Sprintf("%s-%s", "aqueduct", string(b)), nil
}

func fetchGCPCredential(
	ctx context.Context,
	engineResource *models.Resource,
	vaultObject vault.Vault,
) (*shared.GCPConfig, error) {
	config, err := auth.ReadConfigFromSecret(ctx, engineResource.ID, vaultObject)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read integration config from vault.")
	}

	k8sConfig, err := lib_utils.ParseK8sConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse Kubernetes config")
	}

	return k8sConfig.GCPConfig, nil
}

func fetchAWSCredential(
	ctx context.Context,
	engineResource *models.Resource,
	vaultObject vault.Vault,
) (*shared.AWSConfig, error) {
	if _, ok := engineResource.Config[shared.K8sCloudResourceIdKey]; !ok {
		return nil, errors.New("No cloud resource ID found in the engine resource object.")
	}
	cloudResourceID, err := uuid.Parse(engineResource.Config[shared.K8sCloudResourceIdKey])
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse cloud resource ID")
	}

	config, err := auth.ReadConfigFromSecret(ctx, cloudResourceID, vaultObject)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read cloud resource config from vault.")
	}

	awsConfig, err := lib_utils.ParseAWSConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse AWS config.")
	}

	return awsConfig, nil
}

// SetupTerraformDirectory copies all files and folders in the Terraform template directory to the
// cloud integration's destination directory, which is ~/.aqueduct/server/cloud_integration/<name>/eks.
func SetupTerraformDirectory(src, dst string) error {
	// Create the destination directory if it doesn't exist.
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}

	_, stdErr, err := lib_utils.RunCmd(
		"cp",
		[]string{
			"-R", // we could have used -T to not create a directory if the source is a directory, but it's not supported on macOS
			fmt.Sprintf("%s%s.", src, string(filepath.Separator)),
			dst,
		},
		"",
		false,
	)
	if err != nil {
		return errors.New(stdErr)
	}

	return nil
}
