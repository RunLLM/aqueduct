package engine

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultAwsRegion = "us-east-2"
)

func waitForInProgressOperators(
	ctx context.Context,
	inProgressOps map[uuid.UUID]operator.Operator,
	pollInterval time.Duration,
	timeout time.Duration,
) {
	start := time.Now()
	for len(inProgressOps) > 0 {
		if time.Since(start) > timeout {
			return
		}

		for opID, op := range inProgressOps {
			execState, err := op.Poll(ctx)

			// Resolve any jobs that aren't actively running or failed. We don't are if they succeeded or failed,
			// since this is called after engestration exits.
			if err != nil || execState.Status != shared.RunningExecutionStatus {
				delete(inProgressOps, opID)
			}
		}
		time.Sleep(pollInterval)
	}
}

func opFailureError(failureType shared.FailureType, op operator.Operator) error {
	if failureType == shared.SystemFailure {
		return ErrOpExecSystemFailure
	} else if failureType == shared.UserFatalFailure {
		log.Errorf("Failed due to user error. Operator name %s, id %s.", op.Name(), op.ID())
		return ErrOpExecBlockingUserFailure
	}
	return errors.Newf("Internal error: Unsupported failure type %v", failureType)
}

// We should only stop orchestration on system or fatal user errors.
func shouldStopExecution(execState *shared.ExecutionState) bool {
	return execState.Status == shared.FailedExecutionStatus && *execState.FailureType != shared.UserNonFatalFailure
}

func generateJobManagerConfig(
	ctx context.Context,
	dag *models.DAG,
	aqPath string,
	vault vault.Vault,
) (job.Config, error) {
	switch dag.EngineConfig.Type {
	case shared.AqueductEngineType:
		return &job.ProcessConfig{
			BinaryDir:          path.Join(aqPath, job.BinaryDir),
			OperatorStorageDir: path.Join(aqPath, job.OperatorStorageDir),
		}, nil
	case shared.K8sEngineType:
		if dag.StorageConfig.Type != shared.S3StorageType && dag.StorageConfig.Type != shared.GCSStorageType {
			return nil, errors.New("Must use S3 or GCS storage config for K8s engine.")
		}

		var awsAccessKeyId, awsSecretAccessKey string
		if dag.StorageConfig.Type == shared.S3StorageType {
			keyId, secretKey, err := extractAwsCredentials(dag.StorageConfig.S3Config)
			if err != nil {
				return nil, errors.Wrap(err, "Unable to extract AWS credentials from file.")
			}

			awsAccessKeyId = keyId
			awsSecretAccessKey = secretKey
		}

		k8sIntegrationId := dag.EngineConfig.K8sConfig.IntegrationId
		config, err := auth.ReadConfigFromSecret(ctx, k8sIntegrationId, vault)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to read config from vault.")
		}
		k8sConfig, err := ParseK8sConfig(config)
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
		if dag.StorageConfig.Type != shared.S3StorageType {
			return nil, errors.New("Must use S3 for Lambda engine.")
		}
		lambdaIntegrationId := dag.EngineConfig.LambdaConfig.IntegrationId
		config, err := auth.ReadConfigFromSecret(ctx, lambdaIntegrationId, vault)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to read config from vault.")
		}
		lambdaConfig, err := ParseLambdaConfig(config)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to get integration.")
		}

		var awsAccessKeyId, awsSecretAccessKey string
		if dag.StorageConfig.Type == shared.S3StorageType {
			keyId, secretKey, err := extractAwsCredentials(dag.StorageConfig.S3Config)
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
	default:
		return nil, errors.New("Unsupported engine type.")
	}
}

// ParseK8sConfig takes in an auth.Config and parses into a K8s config.
// It also returns an error, if any.
func ParseK8sConfig(conf auth.Config) (*integration.K8sIntegrationConfig, error) {
	data, err := conf.Marshal()
	if err != nil {
		return nil, err
	}

	var c integration.K8sIntegrationConfig
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

func ParseLambdaConfig(conf auth.Config) (*integration.LambdaIntegrationConfig, error) {
	data, err := conf.Marshal()
	if err != nil {
		return nil, err
	}

	var c integration.LambdaIntegrationConfig
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
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
