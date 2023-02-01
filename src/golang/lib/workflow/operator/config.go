package operator

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
)

const (
	DefaultAwsRegion = "us-east-2"
)

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

func GenerateJobManagerConfig(
	ctx context.Context,
	engineConfig shared.EngineConfig,
	storageConfig *shared.StorageConfig,
	aqPath string,
	vault vault.Vault,
) (job.Config, error) {
	switch engineConfig.Type {
	case shared.AqueductEngineType:
		return &job.ProcessConfig{
			BinaryDir:          path.Join(aqPath, job.BinaryDir),
			OperatorStorageDir: path.Join(aqPath, job.OperatorStorageDir),
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

		k8sIntegrationId := engineConfig.K8sConfig.IntegrationId
		config, err := auth.ReadConfigFromSecret(ctx, k8sIntegrationId, vault)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to read config from vault.")
		}
		k8sConfig, err := lib_utils.ParseK8sConfig(config)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to get integration.")
		}
		return &job.K8sJobManagerConfig{
			KubeconfigPath:     k8sConfig.KubeconfigPath,
			ClusterName:        k8sConfig.ClusterName,
			UseSameCluster:     bool(k8sConfig.UseSameCluster),
			AwsAccessKeyId:     awsAccessKeyId,
			AwsSecretAccessKey: awsSecretAccessKey,
			AwsRegion:          DefaultAwsRegion,
		}, nil
	case shared.LambdaEngineType:
		if storageConfig.Type != shared.S3StorageType {
			return nil, errors.New("Must use S3 for Lambda engine.")
		}
		lambdaIntegrationId := engineConfig.LambdaConfig.IntegrationId
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

		return &job.LambdaJobManagerConfig{
			RoleArn:            lambdaConfig.RoleArn,
			AwsAccessKeyId:     awsAccessKeyId,
			AwsSecretAccessKey: awsSecretAccessKey,
		}, nil
	case shared.DatabricksEngineType:
		if storageConfig.Type != shared.S3StorageType {
			return nil, errors.New("Must use S3 storage config for Databricks engine.")
		}
		databricksIntegrationId := engineConfig.DatabricksConfig.IntegrationId
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
		return &job.DatabricksJobManagerConfig{
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
