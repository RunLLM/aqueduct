package engine

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"time"

	db_artifact "github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
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
			execState, err := op.GetExecState(ctx)

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

func convertToPreviewArtifactResponse(ctx context.Context, artf artifact.Artifact) (*PreviewArtifactResults, error) {
	content, err := artf.GetContent(ctx)
	if err != nil {
		return nil, err
	}

	if artf.Type() == db_artifact.FloatType {
		val, err := strconv.ParseFloat(string(content), 32)
		if err != nil {
			return nil, err
		}

		return &PreviewArtifactResults{
			Metric: &previewFloatArtifactResponse{
				Val: val,
			},
		}, nil
	} else if artf.Type() == db_artifact.BoolType {
		passed, err := strconv.ParseBool(string(content))
		if err != nil {
			return nil, err
		}

		return &PreviewArtifactResults{
			Check: &previewBoolArtifactResponse{
				Passed: passed,
			},
		}, nil
	} else if artf.Type() == db_artifact.JsonType {
		return &PreviewArtifactResults{
			Param: &previewParamArtifactResponse{
				Val: string(content),
			},
		}, nil
	} else if artf.Type() == db_artifact.TableType {
		metadata, err := artf.GetMetadata(ctx)
		if err != nil {
			metadata = &artifact_result.Metadata{}
		}
		return &PreviewArtifactResults{
			Table: &previewTableArtifactResponse{
				TableSchema: metadata.Schema,
				Data:        string(content),
			},
		}, nil
	}
	return nil, errors.Newf("Unsupported artifact type %s", artf.Type())
}

func generateJobManagerConfig(
	ctx context.Context,
	dbWorkflowDag *workflow_dag.DBWorkflowDag,
	aqPath string,
	vault vault.Vault,
) (job.Config, error) {
	switch dbWorkflowDag.EngineConfig.Type {
	case shared.AqueductEngineType:
		return &job.ProcessConfig{
			BinaryDir:          path.Join(aqPath, job.BinaryDir),
			OperatorStorageDir: path.Join(aqPath, job.OperatorStorageDir),
		}, nil
	case shared.K8sEngineType:
		if dbWorkflowDag.StorageConfig.Type != shared.S3StorageType {
			return nil, errors.New("Must use S3 storage config for K8s engine.")
		}
		awsAccessKeyId, awsSecretAccessKey, err := extractAwsCredentials(dbWorkflowDag.StorageConfig.S3Config)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to extract AWS credentials from file.")
		}

		k8sIntegrationId := dbWorkflowDag.EngineConfig.K8sConfig.IntegrationId
		config, err := auth.ReadConfigFromSecret(ctx, k8sIntegrationId, vault)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to read config from vault.")
		}
		k8sConfig, err := parseConfig(config)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to get integration.")
		}
		return &job.K8sJobManagerConfig{
			KubeconfigPath:                   k8sConfig.KubeconfigPath,
			ClusterName:                      k8sConfig.ClusterName,
			AwsAccessKeyId:                   awsAccessKeyId,
			AwsSecretAccessKey:               awsSecretAccessKey,
			AwsRegion:                        DefaultAwsRegion,
			FunctionDockerImage:              DefaultFunctionDockerImage,
			ParameterDockerImage:             DefaultParameterDockerImage,
			PostgresConnectorDockerImage:     DefaultPostgresConnectorDockerImage,
			SnowflakeConnectorDockerImage:    DefaultSnowflakeConnectorDockerImage,
			MySqlConnectorDockerImage:        DefaultMySqlConnectorDockerImage,
			SqlServerConnectorDockerImage:    DefaultSqlServerConnectorDockerImage,
			BigQueryConnectorDockerImage:     DefaultBigQueryConnectorDockerImage,
			GoogleSheetsConnectorDockerImage: DefaultGoogleSheetsConnectorDockerImage,
			SalesforceConnectorDockerImage:   DefaultSalesforceConnectorDockerImage,
			S3ConnectorDockerImage:           DefaultS3ConnectorDockerImage,
		}, nil
	default:
		return nil, errors.New("Unsupported engine type.")
	}
}

// parseConfig takes in an auth.Config and parses into a config.
// It also returns an error, if any.
func parseConfig(conf auth.Config) (*shared.K8sIntegrationConfig, error) {
	data, err := conf.Marshal()
	if err != nil {
		return nil, err
	}

	var c shared.K8sIntegrationConfig
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
			awsAccessKeyId = fileScanner.Text()
			if fileScanner.Scan() {
				fmt.Sscanf(fileScanner.Text(), "aws_access_key_id = %v", &awsAccessKeyId)
			} else {
				return "", "", errors.New("Unable to extract AWS credentials.")
			}
			if fileScanner.Scan() {
				fmt.Sscanf(fileScanner.Text(), "aws_secret_access_key = %v", &awsSecretAccessKey)
			} else {
				return "", "", errors.New("Unable to extract AWS credentials.")
			}

			return awsAccessKeyId, awsSecretAccessKey, nil
		}
	}
	return "", "", errors.New("Unable to extract AWS credentials.")
}
