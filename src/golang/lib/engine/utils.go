package engine

import (
	"context"
	"path"
	"strconv"
	"time"

	db_artifact "github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
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
	log.Info("StdErr: ")
	log.Info(execState.UserLogs.StdErr)
	log.Info("StdOut: ")
	log.Info(execState.UserLogs.Stdout)
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

func generateJobManagerConfig(dbWorkflowDag *workflow_dag.DBWorkflowDag, aqPath string) (job.Config, error) {
	switch dbWorkflowDag.EngineConfig.Type {
	case shared.AqueductEngineType:
		return &job.ProcessConfig{
			BinaryDir:          path.Join(aqPath, job.BinaryDir),
			OperatorStorageDir: path.Join(aqPath, job.OperatorStorageDir),
		}, nil
	case shared.K8sEngineType:
		return &job.K8sConfig{
			KubeConfigPath:                   "/home/ubuntu/.kube/config",
			AwsRegion:                        "us-east-2",
			ClusterName:                      "aqueduct-hari",
			AwsAccessKeyId:                   "",
			AwsSecretAccessKey:               "",
			FunctionDockerImage:              "aqueducthq/function",
			ParameterDockerImage:             "aqueducthq/param",
			PostgresConnectorDockerImage:     "aqueducthq/postgres-connector",
			SnowflakeConnectorDockerImage:    "aqueducthq/snowflake-connector",
			MySqlConnectorDockerImage:        "aqueducthq/mysql-connector",
			SqlServerConnectorDockerImage:    "aqueducthq/sqlserver-connector",
			BigQueryConnectorDockerImage:     "aqueducthq/bigquery-connector",
			GoogleSheetsConnectorDockerImage: "aqueducthq/googlesheets-connector",
			SalesforceConnectorDockerImage:   "aqueducthq/salesforce-connector",
			S3ConnectorDockerImage:           "aqueducthq/s3-connector",
		}, nil
	default:
		return nil, errors.New("Unsupported engine type.")
	}
}
