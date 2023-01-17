package job

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	databricks_lib "github.com/aqueducthq/aqueduct/lib/databricks"
	databricks_sdk "github.com/databricks/databricks-sdk-go"
	"github.com/databricks/databricks-sdk-go/service/jobs"
	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
)

type DatabricksJobManager struct {
	databricksClient *databricks_sdk.WorkspaceClient
	conf             *DatabricksJobManagerConfig
	runMap           map[string]int64
}

func NewDatabricksJobManager(conf *DatabricksJobManagerConfig) (*DatabricksJobManager, error) {
	databricksClient, err := databricks_lib.NewWorkspaceClient(
		conf.WorkspaceURL,
		conf.AccessToken,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create Databricks Workspace Client.")
	}

	return &DatabricksJobManager{
		databricksClient: databricksClient,
		conf:             conf,
		runMap:           map[string]int64{},
	}, nil
}

func (j *DatabricksJobManager) mapJobTypeToFile(spec Spec) (string, string, error) {
	if spec.Type() == FunctionJobType {
		functionSpec, ok := spec.(*FunctionSpec)
		if !ok {
			return "", "", ErrInvalidJobSpec
		}

		functionSpec.FunctionExtractPath = defaultFunctionExtractPath
		storageConfig, err := spec.GetStorageConfig()
		if err != nil {
			return "", "", errors.Wrap(err, "Spec unexpectedly has no storage config.")
		}
		storageConfig.S3Config.AWSAccessKeyID = j.conf.AwsAccessKeyID
		storageConfig.S3Config.AWSSecretAccessKey = j.conf.AwsSecretAccessKey

		specStr, err := EncodeSpec(spec, JsonSerializationType)
		if err != nil {
			return "", "", errors.Wrap(err, "Unable to encode spec.")
		}
		return databricks_lib.DatabricksFunctionScript, specStr, nil

	} else if spec.Type() == ParamJobType {
		specStr, err := EncodeSpec(spec, JsonSerializationType)
		if err != nil {
			return "", "", errors.Wrap(err, "Unable to encode spec.")
		}
		return databricks_lib.DatabricksParamScript, specStr, nil

	} else if spec.Type() == AuthenticateJobType ||
		spec.Type() == LoadJobType ||
		spec.Type() == ExtractJobType ||
		spec.Type() == LoadTableJobType ||
		spec.Type() == DeleteSavedObjectsJobType ||
		spec.Type() == DiscoverJobType {
		specStr, err := EncodeSpec(spec, JsonSerializationType)
		if err != nil {
			return "", "", errors.Wrap(err, "Unable to encode spec.")
		}
		return databricks_lib.DatabricksDataScript, specStr, nil

	} else if spec.Type() == SystemMetricJobType {
		specStr, err := EncodeSpec(spec, JsonSerializationType)
		if err != nil {
			return "", "", errors.Wrap(err, "Unable to encode spec.")
		}
		return databricks_lib.DatabricksMetricScript, specStr, nil

	} else {
		return "", "", errors.New("Unsupported JobType was passed in.")
	}
}

func (j *DatabricksJobManager) Config() Config {
	return j.conf
}

func (j *DatabricksJobManager) Launch(
	ctx context.Context,
	name string,
	spec Spec,
) JobError {
	log.Infof("Running %s job %s.", spec.Type(), name)

	scriptFile, specStr, err := j.mapJobTypeToFile(spec)
	if err != nil {
		return systemError(err)
	}
	storageConfig, err := spec.GetStorageConfig()
	if err != nil {
		return systemError(errors.Wrap(err, "Spec unexpectedly has no storage config."))
	}
	bucket := storageConfig.S3Config.Bucket
	pythonFilePath := fmt.Sprintf("%s/%s", bucket, scriptFile)

	jobID, err := databricks_lib.CreateJob(ctx, j.databricksClient, name, j.conf.S3InstanceProfileARN, pythonFilePath)
	if err != nil {
		return systemError(errors.Wrap(err, "Error creating job in Databricks."))
	}
	runID, err := databricks_lib.RunNow(ctx, j.databricksClient, jobID, specStr)
	if err != nil {
		return systemError(errors.Wrap(err, "Error runnning job in Databricks."))
	}
	j.runMap[name] = runID
	return nil
}

func (j *DatabricksJobManager) Poll(ctx context.Context, name string) (shared.ExecutionStatus, JobError) {
	runId, ok := j.runMap[name]
	if !ok {
		return shared.UnknownExecutionStatus, jobMissingError(errors.New("Job doesn't exist."))
	}

	getRunReq := &jobs.GetRun{
		RunId: runId,
	}
	getRunResp, err := j.databricksClient.Jobs.GetRun(ctx, *getRunReq)
	if err != nil {
		return shared.UnknownExecutionStatus, systemError(errors.Wrap(err, "Unable to get run from databricks."))
	}

	switch getRunResp.State.LifeCycleState {
	case jobs.RunLifeCycleStatePending, jobs.RunLifeCycleStateRunning, jobs.RunLifeCycleStateTerminating:
		return shared.RunningExecutionStatus, nil
	case jobs.RunLifeCycleStateInternalError:
		return shared.FailedExecutionStatus, nil
	case jobs.RunLifeCycleStateTerminated:
		switch getRunResp.State.ResultState {
		case jobs.RunResultStateSuccess:
			return shared.SucceededExecutionStatus, nil
		default:
			return shared.FailedExecutionStatus, nil
		}
	default:
		return shared.UnknownExecutionStatus, noopError(errors.New("Unable to determine job status."))
	}
}

func (j *DatabricksJobManager) DeployCronJob(
	ctx context.Context,
	name string,
	period string,
	spec Spec,
) JobError {
	return nil
}

func (j *DatabricksJobManager) CronJobExists(ctx context.Context, name string) bool {
	return false
}

func (j *DatabricksJobManager) EditCronJob(ctx context.Context, name string, cronString string) JobError {
	return nil
}

func (j *DatabricksJobManager) DeleteCronJob(ctx context.Context, name string) JobError {
	return nil
}
