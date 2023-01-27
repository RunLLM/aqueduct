package engine

import (
	"context"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/job"
	dag_utils "github.com/aqueducthq/aqueduct/lib/workflow/dag"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/databricks/databricks-sdk-go/service/jobs"
	"github.com/dropbox/godropbox/errors"
)

// We seperate out the execution step for Databricks Jobs since
// Databricks takes care of launching Tasks.
// Steps:
// 1. Convert each operator into a Task (includes parent dependency).
// 2. Create a multi-task Job with all the previously created Tasks.
// 3. Launch job asynchronously
// 4. Poll on each Task and update accordingly.
func ExecuteDatabricks(
	ctx context.Context,
	dag dag_utils.WorkflowDag,
	workflowName string,
	workflowRunMetadata *WorkflowRunMetadata,
	timeConfig *AqueductTimeConfig,
	opExecMode operator.ExecutionMode,
	databricksJobManager *job.DatabricksJobManager,
) error {
	inProgressOps := workflowRunMetadata.InProgressOps
	completedOps := workflowRunMetadata.CompletedOps

	// Convert the operators into tasks
	taskList, err := CreateTaskList(ctx, dag, workflowName, databricksJobManager)
	if err != nil {
		return errors.Wrap(err, "Unable to convert operators to Databricks tasks.")
	}

	// Launch the workflow job with all tasks
	_, err = databricksJobManager.LaunchMultipleTaskJob(ctx, workflowName, taskList)
	if err != nil {
		return errors.Wrap(err, "Unable to launch workflow job on Databricks.")
	}

	for _, op := range dag.Operators() {
		inProgressOps[op.ID()] = op
	}

	if len(inProgressOps) == 0 {
		return errors.Newf("No initial operators to schedule.")
	}

	// Wait a little bit for all active operators to finish before exiting on failure.
	defer waitForInProgressOperators(ctx, inProgressOps, timeConfig.OperatorPollInterval, timeConfig.CleanupTimeout)

	start := time.Now()
	var operatorError error

	for len(inProgressOps) > 0 {
		if time.Since(start) > timeConfig.ExecTimeout {
			return errors.New("Reached timeout waiting for workflow to complete.")
		}

		for _, op := range inProgressOps {
			// Poll on the individual operator
			executionStatus, err := databricksJobManager.Poll(ctx, op.JobSpec().JobName())
			op.UpdateExecState(&shared.ExecutionState{Status: executionStatus})
			execState := op.ExecState()
			if err != nil {
				return err
			}

			if !execState.Terminated() {
				continue
			}

			// From here on we can assume that the operator has terminated.
			if opExecMode == operator.Publish {
				err := op.PersistResult(ctx)
				if err != nil {
					return errors.Wrapf(err, "Error when finishing execution of operator %s", op.Name())
				}
			}
			// Capture the first failed operator.

			if shouldStopExecution(execState) && operatorError == nil {
				operatorError = opFailureError(*execState.FailureType, op)
			}

			// Add the operator to the completed stack, and remove it from the in-progress one.
			if _, ok := completedOps[op.ID()]; ok {
				return errors.Newf("Internal error: operator %s was completed twice.", op.Name())
			}
			completedOps[op.ID()] = op
			delete(inProgressOps, op.ID())
		}
		time.Sleep(timeConfig.OperatorPollInterval)
	}

	if len(completedOps) != len(dag.Operators()) {
		return errors.Newf("Internal error: %d operators were provided but only %d completed.", len(dag.Operators()), len(completedOps))
	}
	if operatorError != nil {
		return operatorError
	}
	return nil
}

// Takes in a DAG and converts the operators into tasks within a Datbricks Job.
func CreateTaskList(
	ctx context.Context,
	workflowDag dag_utils.WorkflowDag,
	workflowName string,
	databricksJobManager *job.DatabricksJobManager,
) ([]jobs.JobTaskSettings, error) {
	dag := workflowDag
	var taskList []jobs.JobTaskSettings

	for _, op := range dag.Operators() {

		// Get the upstream dependent operators
		parentOperators, err := dag.OperatorParents(op)
		parentOperatorNames := make([]string, 0, len(parentOperators))
		for _, op := range parentOperators {
			parentOperatorNames = append(parentOperatorNames, op.JobSpec().JobName())
		}
		if err != nil {
			return nil, errors.Wrap(err, "Unable to get operator parents.")
		}

		task, err := databricksJobManager.CreateTask(
			ctx,
			workflowName,
			op.JobSpec(),
			parentOperatorNames,
		)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to create task from operator.")
		}
		taskList = append(taskList, *task)
	}
	return taskList, nil
}
