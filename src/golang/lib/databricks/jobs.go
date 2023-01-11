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
	pythonFilePath string,
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
		Tasks: []jobs.JobTaskSettings{
			{
				TaskKey: name,
				NewCluster: &clusters.CreateCluster{
					SparkVersion: latestLTS,
					NumWorkers:   NumWorkers,
					NodeTypeId:   NodeTypeId,
					AwsAttributes: &clusters.AwsAttributes{
						InstanceProfileArn: s3InstanceProfileArn,
					},
				},
				SparkPythonTask: &jobs.SparkPythonTask{
					PythonFile: pythonFilePath,
				},
				Libraries: []libraries.Library{
					{
						Pypi: &libraries.PythonPyPiLibrary{
							Package: fmt.Sprintf("aqueduct-ml==%s", lib.ServerVersionNumber),
						},
					},
				},
			},
		},
	}
	createResp, err := databricksClient.Jobs.Create(ctx, *createRequest)
	if err != nil {
		return -1, errors.Wrap(err, "Error creating a job in Databricks.")
	}
	return createResp.JobId, nil
}

func RunNow(
	ctx context.Context,
	databricksClient *databricks_sdk.WorkspaceClient,
	jobId int64,
	specStr string,
) (int64, error) {
	runResp, err := databricksClient.Jobs.RunNow(
		ctx,
		jobs.RunNow{
			JobId:        jobId,
			PythonParams: []string{"--spec", specStr},
		},
	)
	if err != nil {
		return -1, errors.Wrap(err, "Error launching job in Databricks.")
	}
	return runResp.RunId, nil
}
