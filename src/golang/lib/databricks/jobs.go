package databricks

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib"
	databricks_sdk "github.com/databricks/databricks-sdk-go"
	"github.com/databricks/databricks-sdk-go/service/clusters"
	"github.com/databricks/databricks-sdk-go/service/jobs"
	"github.com/databricks/databricks-sdk-go/service/libraries"
	"github.com/dropbox/godropbox/errors"
)

func NewWorkspaceClient(
	workspaceUrl string,
	accessToken string,
) (*databricks_sdk.WorkspaceClient, error) {
	dConfig := &databricks_sdk.Config{
		Host:  workspaceUrl,
		Token: accessToken,
	}
	datatbricksClient, err := databricks_sdk.NewWorkspaceClient(dConfig)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create Databricks client.")
	}
	return datatbricksClient, nil
}

func ListJobs(
	ctx context.Context,
	databricksClient *databricks_sdk.WorkspaceClient,
) ([]jobs.Job, error) {
	jobs, err := databricksClient.Jobs.ListAll(
		ctx,
		jobs.List{},
	)
	if err != nil {
		return nil, errors.Wrap(err, "Error launching job in Databricks.")
	}
	return jobs, nil
}

func CreateJob(
	ctx context.Context,
	databricksClient *databricks_sdk.WorkspaceClient,
	name string,
	s3InstanceProfileArn string,
	tasks []jobs.JobTaskSettings,
) (int64, error) {

	sparkVersions, err := databricksClient.Clusters.SparkVersions(ctx)
	if err != nil {
		return -1, errors.Wrap(err, "Error selecting a spark version.")
	}
	// Select the latest LTS version.
	latestLTS, err := sparkVersions.Select(clusters.SparkVersionRequest{
		Latest:          true,
		LongTermSupport: true,
	})
	if err != nil {
		return -1, errors.Wrap(err, "Error selecting a spark version.")
	}

	createRequest := &jobs.CreateJob{
		Name: name,
		JobClusters: []jobs.JobCluster{
			{
				JobClusterKey: workflowNameToJobClusterKey(name),
				NewCluster: &clusters.CreateCluster{
					SparkVersion: latestLTS,
					NumWorkers:   NumWorkers,
					NodeTypeId:   NodeTypeId,
					AwsAttributes: &clusters.AwsAttributes{
						InstanceProfileArn: s3InstanceProfileArn,
					},
				},
			},
		},
		Tasks: tasks,
	}
	createResp, err := databricksClient.Jobs.Create(ctx, *createRequest)
	if err != nil {
		return -1, errors.Wrap(err, "Error creating a job in Databricks.")
	}
	return createResp.JobId, nil
}

func CreateTask(
	ctx context.Context,
	databricksClient *databricks_sdk.WorkspaceClient,
	workflowName string,
	name string,
	upstreamTaskNames []string,
	pythonFilePath string,
	specStr string,
) (*jobs.JobTaskSettings, error) {

	jobClusterKey := workflowNameToJobClusterKey(workflowName)

	taskDependenciesList := make([]jobs.TaskDependenciesItem, 0, len(upstreamTaskNames))
	for _, taskName := range upstreamTaskNames {
		taskDependenciesList = append(taskDependenciesList, jobs.TaskDependenciesItem{TaskKey: taskName})
	}

	task := &jobs.JobTaskSettings{
		TaskKey:       name,
		JobClusterKey: jobClusterKey,
		DependsOn:     taskDependenciesList,
		SparkPythonTask: &jobs.SparkPythonTask{
			PythonFile: pythonFilePath,
			Parameters: []string{"--spec", specStr},
		},
		Libraries: []libraries.Library{
			{
				Pypi: &libraries.PythonPyPiLibrary{
					Package: fmt.Sprintf("aqueduct-ml==%v", lib.ServerVersionNumber),
				},
			},
			{
				Pypi: &libraries.PythonPyPiLibrary{
					Package: "snowflake-sqlalchemy",
				},
			},
		},
	}
	return task, nil
}

func RunNow(
	ctx context.Context,
	databricksClient *databricks_sdk.WorkspaceClient,
	jobID int64,
) (int64, error) {
	runResp, err := databricksClient.Jobs.RunNow(
		ctx,
		jobs.RunNow{
			JobId: jobID,
		},
	)
	if err != nil {
		return -1, errors.Wrap(err, "Error launching job in Databricks.")
	}
	return runResp.RunId, nil
}

func GetRun(
	ctx context.Context,
	databricksClient *databricks_sdk.WorkspaceClient,
	runID int64,
) (*jobs.Run, error) {
	getRunReq := &jobs.GetRun{
		RunId: runID,
	}
	getRunResp, err := databricksClient.Jobs.GetRun(ctx, *getRunReq)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get run from databricks.")
	}
	return getRunResp, nil
}

func GetTaskRunIDs(
	ctx context.Context,
	databricksClient *databricks_sdk.WorkspaceClient,
	runID int64,
) (map[string]int64, error) {
	taskNameToID := make(map[string]int64)
	runResp, err := GetRun(ctx, databricksClient, runID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get run metadata.")
	}
	for _, taskResp := range runResp.Tasks {
		taskNameToID[taskResp.TaskKey] = taskResp.RunId
	}
	return taskNameToID, nil
}

func workflowNameToJobClusterKey(workflowName string) string {
	return fmt.Sprintf("%s_cluster", workflowName)
}
