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

func (j *DatabricksJobManager) Config() Config {
	return j.conf
}

func (j *DatabricksJobManager) Launch(
	ctx context.Context,
	name string,
	spec Spec,
) JobError {
	log.Infof("Running %s job %s.", spec.Type(), name)

	task, err := j.CreateTask(ctx, name, spec, []string{})
	if err != nil {
		return systemError(err)
	}

	jobID, err := databricks_lib.CreateJob(ctx, j.databricksClient, name, j.conf.S3InstanceProfileARN, []jobs.JobTaskSettings{*task})
	if err != nil {
		return systemError(errors.Wrap(err, "Error creating job in Databricks."))
	}
	runID, err := databricks_lib.RunNow(ctx, j.databricksClient, jobID)
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

	runResp, err := databricks_lib.GetRun(ctx, j.databricksClient, runId)
	if err != nil {
		return shared.UnknownExecutionStatus, systemError(errors.Wrap(err, "Unable to get run from databricks."))
	}

	switch runResp.State.LifeCycleState {
	case "BLOCKED":
		return shared.PendingExecutionStatus, nil
	case jobs.RunLifeCycleStatePending, jobs.RunLifeCycleStateRunning, jobs.RunLifeCycleStateTerminating:
		return shared.RunningExecutionStatus, nil
	case jobs.RunLifeCycleStateSkipped:
		return shared.CanceledExecutionStatus, nil
	case jobs.RunLifeCycleStateInternalError:
		return shared.FailedExecutionStatus, nil
	case jobs.RunLifeCycleStateTerminated:
		switch runResp.State.ResultState {
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

func (j *DatabricksJobManager) CreateTask(
	ctx context.Context,
	workflowName string,
	spec Spec,
	parentOperatorNames []string,
) (*jobs.JobTaskSettings, error) {
	// Get the entrypoint file for Databricks.
	storageConfig, err := spec.GetStorageConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Spec unexpectedly has no storage config.")
	}
	scriptFile, specStr, err := j.mapJobTypeToFile(spec)
	if err != nil {
		return nil, err
	}
	bucket := storageConfig.S3Config.Bucket
	pythonFilePath := fmt.Sprintf("%s/%s", bucket, scriptFile)

	task, err := databricks_lib.CreateTask(
		ctx,
		j.databricksClient,
		workflowName,
		spec.JobName(),
		parentOperatorNames,
		pythonFilePath,
		specStr,
	)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (j *DatabricksJobManager) LaunchMultipleTaskJob(
	ctx context.Context,
	name string,
	taskList []jobs.JobTaskSettings,
) (int64, JobError) {
	// Create and register the job with Databricks.
	jobID, err := databricks_lib.CreateJob(ctx, j.databricksClient, name, j.conf.S3InstanceProfileARN, taskList)
	if err != nil {
		return -1, systemError(errors.Wrap(err, "Error creating job in Databricks."))
	}
	// Trigger the job to execute asynchronously.
	runID, err := databricks_lib.RunNow(ctx, j.databricksClient, jobID)
	if err != nil {
		return -1, systemError(errors.Wrap(err, "Error runnning job in Databricks."))
	}
	j.runMap[name] = runID
	// Get the runIDs of the individual tasks for polling purposes.
	taskMap, err := databricks_lib.GetTaskRunIDs(ctx, j.databricksClient, runID)
	if err != nil {
		return -1, systemError(errors.Wrap(err, "Error creating the task id map."))
	}
	for taskName, taskID := range taskMap {
		j.runMap[taskName] = taskID
	}
	return runID, nil
}

func (j *DatabricksJobManager) mapJobTypeToFile(spec Spec) (string, string, error) {
	// Add S3 Access Keys to all specs
	storageConfig, err := spec.GetStorageConfig()
	if err != nil {
		return "", "", errors.Wrap(err, "Spec unexpectedly has no storage config.")
	}
	storageConfig.S3Config.AWSAccessKeyID = j.conf.AwsAccessKeyID
	storageConfig.S3Config.AWSSecretAccessKey = j.conf.AwsSecretAccessKey
	if spec.Type() == FunctionJobType {
		functionSpec, ok := spec.(*FunctionSpec)
		if !ok {
			return "", "", ErrInvalidJobSpec
		}

		functionSpec.FunctionExtractPath = defaultFunctionExtractPath
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
