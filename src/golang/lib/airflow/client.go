package airflow

import (
	"context"

	"github.com/apache/airflow-client-go/airflow"
	"github.com/aqueducthq/aqueduct/lib/errors"
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

func (c *client) getDag(dagId string) (*airflow.DAG, error) {
	dag, resp, err := c.apiClient.DAGApi.GetDag(c.ctx, dagId).Execute()
	if err != nil {
		return nil, wrapApiError(err, "GetDag", resp)
	}

	return &dag, nil
}

// getDagRuns returns all of the Airflow DAGRuns for the Airflow DAG specified.
func (c *client) getDagRuns(dagId string) ([]airflow.DAGRun, error) {
	limitPerFetch := 100 // This is the max number of DAG runs that can be returned in each response.
	offset := 0
	var dagRuns []airflow.DAGRun

	// Keep paginating through DAG runs until there are none in response
	for {
		dagRunsResp, resp, err := c.apiClient.DAGRunApi.GetDagRuns(
			c.ctx,
			dagId,
		).
			OrderBy("start_date").
			Limit(int32(limitPerFetch)).
			Offset(int32(offset)).
			Execute()
		if err != nil {
			return nil, wrapApiError(err, "GetDagRuns", resp)
		}

		if len(*dagRunsResp.DagRuns) == 0 {
			// There are no more DAG Runs
			break
		}

		dagRuns = append(dagRuns, *dagRunsResp.DagRuns...)
		offset += limitPerFetch
	}

	return dagRuns, nil
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

// isDAGPaused returns whether or not the specified DAG is paused.
func (c *client) isDAGPaused(dagID string) (bool, error) {
	dag, err := c.getDag(dagID)
	if err != nil {
		return false, err
	}

	return dag.GetIsPaused(), nil
}

// trigerDAGRun triggers a new DAGRun for the dag specified.
// It first ensures that the DAG is not paused.
func (c *client) triggerDAGRun(dagID string) error {
	// Check if DAG is paused
	paused, err := c.isDAGPaused(dagID)
	if err != nil {
		return err
	}

	if paused {
		// TODO ENG-3002: Automatically unpause the DAG instead of throwing an error
		return errors.Newf("Unable to trigger a new DAG run for %v because it is currently paused. You must unpause it first!", dagID)
	}

	request := c.apiClient.DAGRunApi.PostDagRun(c.ctx, dagID)
	// The PostDagRun API requires the request to have a DAGRun initialized
	request = request.DAGRun(*airflow.NewDAGRunWithDefaults())
	_, _, err = request.Execute()
	return err
}
