package dynamic

import (
	"context"
	"crypto/rand"
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
)

type k8sClusterActionType string

const (
	K8sClusterCreateAction k8sClusterActionType = "create"
	K8sClusterUpdateAction k8sClusterActionType = "update"
)

const (
	stateLockErrMsg          = "Error acquiring the state lock"
	K8sIntegrationNameSuffix = "aqueduct_ondemand_k8s"
)

var TerraformTemplateDir = filepath.Join(os.Getenv("HOME"), ".aqueduct", "server", "template", "aws", "eks")

// PrepareCluster blocks until the cluster is in status "Active".
func PrepareCluster(
	ctx context.Context,
	configDelta *shared.DynamicK8sConfig,
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
				K8sClusterCreateAction,
				engineIntegration,
				integrationRepo,
				vaultObject,
				db,
			)
		} else if engineIntegration.Config[shared.K8sStatusKey] == string(shared.K8sClusterActiveStatus) {
			if len(configDelta.ToMap()) == 0 {
				log.Info("Kubernetes cluster is currently active, proceeding...")
				return nil
			} else {
				log.Info("Kubernetes cluster is currently active, updating the cluster since a non-empty config delta is provided...")
				return CreateOrUpdateK8sCluster(
					ctx,
					configDelta,
					K8sClusterUpdateAction,
					engineIntegration,
					integrationRepo,
					vaultObject,
					db,
				)
			}
		} else {
			engineIntegration, err = PollClusterStatus(ctx, engineIntegration, integrationRepo, vaultObject, db)
			if err != nil {
				return err
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
	configDelta *shared.DynamicK8sConfig,
	action k8sClusterActionType, // can either be k8sClusterCreateAction or k8sClusterUpdateAction
	engineIntegration *models.Integration,
	integrationRepo repos.Integration,
	vaultObject vault.Vault,
	db database.Database,
) error {
	if !(action == K8sClusterCreateAction || action == K8sClusterUpdateAction) {
		return errors.Newf("Unsupport action %s.", action)
	}

	configDeltaMap := configDelta.ToMap()

	if action == K8sClusterUpdateAction && len(configDeltaMap) == 0 {
		return nil // if there is no config delta, we don't need to update anything
	}

	if len(configDeltaMap) > 0 {
		// Update config to reflect the new values.
		for key, value := range configDeltaMap {
			engineIntegration.Config[key] = value
		}
	}

	if err := CheckIfValidConfig(action, engineIntegration.Config); err != nil {
		return err
	}

	var clusterStatus shared.K8sClusterStatusType
	if action == K8sClusterCreateAction {
		clusterStatus = shared.K8sClusterCreatingStatus
	} else {
		clusterStatus = shared.K8sClusterUpdatingStatus
	}

	if err := updateClusterStatus(ctx, clusterStatus, engineIntegration.ID, integrationRepo, db); err != nil {
		return err
	}

	awsConfig, err := fetchAWSCredential(ctx, engineIntegration, vaultObject)
	if err != nil {
		return err
	}

	if err := runTerraformApply(awsConfig, engineIntegration); err != nil {
		return err
	}

	if action == K8sClusterCreateAction {
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
				fmt.Sprintf("AWS_CONFIG_FILE=%s", awsConfig.ConfigFilePath),
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
				engineIntegration.Config[shared.K8sClusterNameKey],
				"--kubeconfig",
				engineIntegration.Config[shared.K8sKubeconfigPathKey],
			),
			engineIntegration.Config[shared.K8sTerraformPathKey],
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
	if err := updateClusterConfig(ctx, action, configDeltaMap, engineIntegration.ID, integrationRepo, db); err != nil {
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
// If skipPodsStatusCheck is set to false, it checks whether there are pods in Running or ContainerCreating
// status and if so, reject the deletion request.
func DeleteK8sCluster(
	ctx context.Context,
	skipPodsStatusCheck bool,
	engineIntegration *models.Integration,
	integrationRepo repos.Integration,
	vaultObject vault.Vault,
	db database.Database,
) error {
	if !skipPodsStatusCheck {
		useSameCluster, err := strconv.ParseBool(engineIntegration.Config[shared.K8sUseSameClusterKey])
		if err != nil {
			return errors.Wrap(err, "Error parsing use_same_cluster flag")
		}

		safe, err := k8s.SafeToDeleteCluster(ctx, useSameCluster, engineIntegration.Config[shared.K8sKubeconfigPathKey])
		if err != nil {
			return err
		}

		if !safe {
			return errors.New("The k8s cluster cannot be deleted because there are pods still running.")
		}
	}

	if err := updateClusterStatus(ctx, shared.K8sClusterTerminatingStatus, engineIntegration.ID, integrationRepo, db); err != nil {
		return err
	}

	// Even for deletion, we need to specify the AWS region, so we need to pass in the actual AWS
	// config instead of a dummy one to generateTerraformVariables.
	awsConfig, err := fetchAWSCredential(ctx, engineIntegration, vaultObject)
	if err != nil {
		return err
	}

	terraformArgs, err := generateTerraformVariables(awsConfig, engineIntegration.Config)
	if err != nil {
		return err
	}

	if _, _, err := lib_utils.RunCmd(
		"terraform",
		append([]string{"destroy", "-auto-approve"}, terraformArgs...),
		engineIntegration.Config[shared.K8sTerraformPathKey],
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
	action k8sClusterActionType,
	configDeltaMap map[string]string,
	engineIntegrationId uuid.UUID,
	integrationRepo repos.Integration,
	db database.Database,
) error {
	if len(configDeltaMap) == 0 {
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
	for key, value := range configDeltaMap {
		engineIntegration.Config[key] = value
	}

	if action == K8sClusterCreateAction {
		// If this is a request to create a new cluster, we need to refresh the desired node counts.
		engineIntegration.Config[shared.K8sDesiredCpuNodeKey] = engineIntegration.Config[shared.K8sMinCpuNodeKey]
		engineIntegration.Config[shared.K8sDesiredGpuNodeKey] = engineIntegration.Config[shared.K8sMinGpuNodeKey]
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
	vaultObject vault.Vault,
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
		engineIntegration.Config[shared.K8sTerraformPathKey],
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
		engineIntegration,
		integrationRepo,
		vaultObject,
		db,
	)
}

func PollClusterStatus(
	ctx context.Context,
	engineIntegration *models.Integration,
	integrationRepo repos.Integration,
	vaultObject vault.Vault,
	db database.Database,
) (*models.Integration, error) {
	if err := ResyncClusterState(ctx, engineIntegration, integrationRepo, vaultObject, db); err != nil {
		return nil, errors.Wrap(err, "Failed to resync cluster state")
	}

	engineIntegration, err := integrationRepo.Get(
		ctx,
		engineIntegration.ID,
		db,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to retrieve engine integration")
	}

	if engineIntegration.Config[shared.K8sStatusKey] == string(shared.K8sClusterTerminatedStatus) {
		// This means the cluster state is resynced to Terminated, so no need to wait.
		return engineIntegration, nil
	}

	log.Infof("Kubernetes cluster is currently in %s status. Waiting for %d seconds before checking again...", engineIntegration.Config[shared.K8sStatusKey], shared.DynamicK8sClusterStatusPollPeriod)
	time.Sleep(shared.DynamicK8sClusterStatusPollPeriod * time.Second)

	engineIntegration, err = integrationRepo.Get(
		ctx,
		engineIntegration.ID,
		db,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to retrieve engine integration")
	}

	return engineIntegration, nil
}

func generateTerraformVariables(
	awsConfig *shared.AWSConfig,
	engineConfig map[string]string,
) ([]string, error) {
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
				fmt.Sprintf("AWS_CONFIG_FILE=%s", awsConfig.ConfigFilePath),
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

	cpuNodeTypeVar := fmt.Sprintf("-var=cpu_node_type=%s", engineConfig[shared.K8sCpuNodeTypeKey])
	gpuNodeTypeVar := fmt.Sprintf("-var=gpu_node_type=%s", engineConfig[shared.K8sGpuNodeTypeKey])
	minCpuNodeVar := fmt.Sprintf("-var=min_cpu_node=%s", engineConfig[shared.K8sMinCpuNodeKey])
	maxCpuNodeVar := fmt.Sprintf("-var=max_cpu_node=%s", engineConfig[shared.K8sMaxCpuNodeKey])
	minGpuNodeVar := fmt.Sprintf("-var=min_gpu_node=%s", engineConfig[shared.K8sMinGpuNodeKey])
	maxGpuNodeVar := fmt.Sprintf("-var=max_gpu_node=%s", engineConfig[shared.K8sMaxGpuNodeKey])

	clusterNameVar := fmt.Sprintf("-var=cluster_name=%s", engineConfig[shared.K8sClusterNameKey])

	return []string{
		accessKeyVar,
		secretAccessKeyVar,
		regionVar,
		credentialPathVar,
		profileVar,
		cpuNodeTypeVar,
		gpuNodeTypeVar,
		minCpuNodeVar,
		maxCpuNodeVar,
		minGpuNodeVar,
		maxGpuNodeVar,
		clusterNameVar,
	}, nil
}

func runTerraformApply(
	awsConfig *shared.AWSConfig,
	engineIntegration *models.Integration,
) error {
	terraformArgs, err := generateTerraformVariables(awsConfig, engineIntegration.Config)
	if err != nil {
		return err
	}

	if _, _, err := lib_utils.RunCmd(
		"terraform",
		append([]string{"apply", "-auto-approve"}, terraformArgs...),
		engineIntegration.Config[shared.K8sTerraformPathKey],
		true,
	); err != nil {
		return errors.Wrap(err, "Terraform apply failed")
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

func fetchAWSCredential(
	ctx context.Context,
	engineIntegration *models.Integration,
	vaultObject vault.Vault,
) (*shared.AWSConfig, error) {
	if _, ok := engineIntegration.Config[shared.K8sCloudIntegrationIdKey]; !ok {
		return nil, errors.New("No cloud integration ID found in the engine integration object.")
	}
	cloudIntegrationId, err := uuid.Parse(engineIntegration.Config[shared.K8sCloudIntegrationIdKey])
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse cloud integration ID")
	}

	config, err := auth.ReadConfigFromSecret(ctx, cloudIntegrationId, vaultObject)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read cloud integration config from vault.")
	}

	awsConfig, err := lib_utils.ParseAWSConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse AWS config.")
	}

	return awsConfig, nil
}
