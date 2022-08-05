package airflow

import (
	"context"

	"github.com/apache/airflow-client-go/airflow"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
)

type client struct {
	apiClient *airflow.APIClient
	ctx       context.Context
}

func newClient(ctx context.Context, authConf auth.Config) (*client, error) {
	conf, err := parseConfig(authConf)
	if err != nil {
		return nil, err
	}

	airflowConf := airflow.NewConfiguration()
	airflowConf.Host = conf.Host
	airflowConf.Scheme = "http"

	apiClient := airflow.NewAPIClient(airflowConf)

	cred := airflow.BasicAuth{
		UserName: conf.Username,
		Password: conf.Password,
	}

	airflowCtx := context.WithValue(ctx, airflow.ContextBasicAuth, cred)

	return &client{
		apiClient: apiClient,
		ctx:       airflowCtx,
	}, nil
}

// getDagRuns returns all of the Airflow DAGRuns for the Airflow DAG specified.
func (c *client) getDagRuns(dagId string) ([]airflow.DAGRun, error) {
	dagRunsResp, resp, err := c.apiClient.DAGRunApi.GetDagRuns(
		c.ctx,
		dagId,
	).Execute()
	if err != nil {
		return nil, wrapApiError(err, "GetDagRuns", resp)
	}

	return *dagRunsResp.DagRuns, nil
}

// getTaskStates returns a map of each taskID to its Airflow TaskState for the Airflow
// DAGRun specified.
func (c *client) getTaskStates(dagId string, dagRunId string) (map[string]airflow.TaskState, error) {
	taskResp, resp, err := c.apiClient.TaskInstanceApi.GetTaskInstances(
		c.ctx,
		dagId,
		dagRunId,
	).Execute()
	if err != nil {
		return nil, wrapApiError(err, "GetTasksInstances", resp)
	}

	taskIdToState := make(map[string]airflow.TaskState, len(*taskResp.TaskInstances))
	for _, task := range *taskResp.TaskInstances {
		taskIdToState[*task.TaskId] = task.GetState()
	}

	return taskIdToState, nil
}
