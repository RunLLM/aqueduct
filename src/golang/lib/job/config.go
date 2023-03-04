package job

import (
	"bufio"
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"path"

	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
)

type ManagerType string

const (
	ProcessType    ManagerType = "process"
	K8sType        ManagerType = "k8s"
	LambdaType     ManagerType = "lambda"
	DatabricksType ManagerType = "databricks"

	DefaultAwsRegion = "us-east-2"
)

type Config interface {
	Type() ManagerType
}

type ProcessConfig struct {
	BinaryDir             string `yaml:"binaryDir" json:"binary_dir"`
	LogsDir               string `yaml:"logsDir" json:"logs_dir"`
	PythonExecutorPackage string `yaml:"pythonExecutorPackage" json:"python_executor_package"`
	OperatorStorageDir    string `yaml:"operatorStorageDir" json:"operator_storage_dir"`
	CondaEnvName          string `yaml:"condaEnvName" json:"conda_env_name"`
}

type K8sJobManagerConfig struct {
	KubeconfigPath     string `yaml:"kubeconfigPath" json:"kubeconfig_path"`
	ClusterName        string `yaml:"clusterName" json:"cluster_name"`
	UseSameCluster     bool   `json:"use_same_cluster"  yaml:"useSameCluster"`
	AwsAccessKeyId     string `yaml:"awsAccessKeyId" json:"aws_access_key_id"`
	AwsSecretAccessKey string `yaml:"awsSecretAccessKey" json:"aws_secret_access_key"`

	// System config, will have defaults
	AwsRegion string `yaml:"awsRegion" json:"aws_region"`

	Dynamic bool `yaml:"dynamic" json:"dynamic"`
}

type LambdaJobManagerConfig struct {
	RoleArn            string `yaml:"roleArn" json:"role_arn"`
	AwsAccessKeyId     string `yaml:"awsAccessKeyId" json:"aws_access_key_id"`
	AwsSecretAccessKey string `yaml:"awsSecretAccessKey" json:"aws_secret_access_key"`
}

type DatabricksJobManagerConfig struct {
	// WorkspaceURL is the full url for the Databricks workspace that
	// Aqueduct operators will run on.
	WorkspaceURL string `yaml:"workspaceUrl" json:"workspace_url"`
	// AccessToken is a Databricks AccessToken for a workspace. Information on how
	// to create tokens can be found here: https://docs.databricks.com/dev-tools/auth.html#personal-access-tokens-for-users
	AccessToken string `yaml:"accessToken" json:"access_token"`
	// Databricks needs an Instance Profile with S3 permissions in order to access metadata
	// storage in S3. Information on how to create this can be found here:
	// https://docs.databricks.com/aws/iam/instance-profile-tutorial.html
	S3InstanceProfileARN string `yaml:"s3InstanceProfileArn" json:"s3_instance_profile_arn"`
	// AWS Access Key ID is passed from the StorageConfig.
	AwsAccessKeyID string `yaml:"awsAccessKeyId" json:"aws_access_key_id"`
	// AWS Secret Access Key is passed from the StorageConfig.
	AwsSecretAccessKey string `yaml:"awsSecretAccessKey" json:"aws_secret_access_key"`
}

func (*ProcessConfig) Type() ManagerType {
	return ProcessType
}

func (*K8sJobManagerConfig) Type() ManagerType {
	return K8sType
}

func (*LambdaJobManagerConfig) Type() ManagerType {
	return LambdaType
}

func (*DatabricksJobManagerConfig) Type() ManagerType {
	return DatabricksType
}

func RegisterGobTypes() {
	gob.Register(&ProcessConfig{})
	gob.Register(&K8sJobManagerConfig{})
	gob.Register(&WorkflowSpec{})
	gob.Register(&WorkflowRetentionSpec{})
	gob.Register(&DynamicTeardownSpec{})
}

func init() {
	RegisterGobTypes()
}

func GenerateNewJobManager(
	ctx context.Context,
	engineConfig shared.EngineConfig,
	storageConfig *shared.StorageConfig,
	aqPath string,
	vault vault.Vault,
) (JobManager, error) {
	jobConfig, err := GenerateJobManagerConfig(
		ctx,
		engineConfig,
		storageConfig,
		aqPath,
		vault,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to generate JobManagerConfig.")
	}

	jobManager, err := NewJobManager(jobConfig)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create JobManager.")
	}
	return jobManager, nil
}

func GenerateJobManagerConfig(
	ctx context.Context,
	engineConfig shared.EngineConfig,
	storageConfig *shared.StorageConfig,
	aqPath string,
	vault vault.Vault,
) (Config, error) {
	switch engineConfig.Type {
	case shared.AqueductEngineType:
		return &ProcessConfig{
			BinaryDir:          path.Join(aqPath, BinaryDir),
			OperatorStorageDir: path.Join(aqPath, OperatorStorageDir),
		}, nil
	case shared.AqueductCondaEngineType:
		return &ProcessConfig{
			BinaryDir:          path.Join(aqPath, BinaryDir),
			OperatorStorageDir: path.Join(aqPath, OperatorStorageDir),
			CondaEnvName:       engineConfig.AqueductCondaConfig.Env,
		}, nil
	case shared.K8sEngineType:
		if storageConfig.Type != shared.S3StorageType && storageConfig.Type != shared.GCSStorageType {
			return nil, errors.New("Must use S3 or GCS storage config for K8s engine.")
		}

		var awsAccessKeyId, awsSecretAccessKey string
		if storageConfig.Type == shared.S3StorageType {
			keyId, secretKey, err := extractAwsCredentials(storageConfig.S3Config)
			if err != nil {
				return nil, errors.Wrap(err, "Unable to extract AWS credentials from file.")
			}

			awsAccessKeyId = keyId
			awsSecretAccessKey = secretKey
		}

		k8sIntegrationId := engineConfig.K8sConfig.IntegrationID
		config, err := auth.ReadConfigFromSecret(ctx, k8sIntegrationId, vault)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to read k8s config from vault.")
		}
		k8sConfig, err := lib_utils.ParseK8sConfig(config)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse k8s config.")
		}
		return &K8sJobManagerConfig{
			KubeconfigPath:     k8sConfig.KubeconfigPath,
			ClusterName:        k8sConfig.ClusterName,
			UseSameCluster:     bool(k8sConfig.UseSameCluster),
			AwsAccessKeyId:     awsAccessKeyId,
			AwsSecretAccessKey: awsSecretAccessKey,
			AwsRegion:          DefaultAwsRegion,
			Dynamic:            bool(k8sConfig.Dynamic),
		}, nil
	case shared.LambdaEngineType:
		if storageConfig.Type != shared.S3StorageType {
			return nil, errors.New("Must use S3 for Lambda engine.")
		}
		lambdaIntegrationId := engineConfig.LambdaConfig.IntegrationID
		config, err := auth.ReadConfigFromSecret(ctx, lambdaIntegrationId, vault)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to read config from vault.")
		}
		lambdaConfig, err := lib_utils.ParseLambdaConfig(config)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to get integration.")
		}

		var awsAccessKeyId, awsSecretAccessKey string
		if storageConfig.Type == shared.S3StorageType {
			keyId, secretKey, err := extractAwsCredentials(storageConfig.S3Config)
			if err != nil {
				return nil, errors.Wrap(err, "Unable to extract AWS credentials from file.")
			}

			awsAccessKeyId = keyId
			awsSecretAccessKey = secretKey
		}

		return &LambdaJobManagerConfig{
			RoleArn:            lambdaConfig.RoleArn,
			AwsAccessKeyId:     awsAccessKeyId,
			AwsSecretAccessKey: awsSecretAccessKey,
		}, nil
	case shared.DatabricksEngineType:
		if storageConfig.Type != shared.S3StorageType {
			return nil, errors.New("Must use S3 storage config for Databricks engine.")
		}
		databricksIntegrationId := engineConfig.DatabricksConfig.IntegrationID
		config, err := auth.ReadConfigFromSecret(ctx, databricksIntegrationId, vault)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to read config from vault.")
		}
		databricksConfig, err := lib_utils.ParseDatabricksConfig(config)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to get integration.")
		}

		var awsAccessKeyId, awsSecretAccessKey string
		if storageConfig.Type == shared.S3StorageType {
			keyId, secretKey, err := extractAwsCredentials(storageConfig.S3Config)
			if err != nil {
				return nil, errors.Wrap(err, "Unable to extract AWS credentials from file.")
			}

			awsAccessKeyId = keyId
			awsSecretAccessKey = secretKey
		}
		return &DatabricksJobManagerConfig{
			WorkspaceURL:         databricksConfig.WorkspaceURL,
			AccessToken:          databricksConfig.AccessToken,
			S3InstanceProfileARN: databricksConfig.S3InstanceProfileARN,
			AwsAccessKeyID:       awsAccessKeyId,
			AwsSecretAccessKey:   awsSecretAccessKey,
		}, nil

	default:
		return nil, errors.New("Unsupported engine type.")
	}
}

func extractAwsCredentials(config *shared.S3Config) (string, string, error) {
	var awsAccessKeyId string
	var awsSecretAccessKey string
	profileString := fmt.Sprintf("[%s]", config.CredentialsProfile)

	file, err := os.Open(config.CredentialsPath)
	if err != nil {
		return "", "", errors.Wrap(err, "Unable to open AWS credentials file.")
	}
	defer file.Close()
	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		if profileString == fileScanner.Text() {
			if fileScanner.Scan() {
				fmt.Sscanf(fileScanner.Text(), "aws_access_key_id=%v", &awsAccessKeyId)
			} else {
				return "", "", errors.New("Unable to extract AWS credentials.")
			}
			if fileScanner.Scan() {
				fmt.Sscanf(fileScanner.Text(), "aws_secret_access_key=%v", &awsSecretAccessKey)
			} else {
				return "", "", errors.New("Unable to extract AWS credentials.")
			}

			return awsAccessKeyId, awsSecretAccessKey, nil
		}
	}
	return "", "", errors.New("Unable to extract AWS credentials.")
}
