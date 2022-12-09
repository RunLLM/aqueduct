package airflow

import (
	"context"
	"time"

	"github.com/apache/airflow-client-go/airflow"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// SyncDAGs syncs all DAGs in dagIDs with any new
// Airflow dag runs since the last sync. It returns an error, if any.
func SyncDAGs(
	ctx context.Context,
	dagIDs []uuid.UUID,
	workflowRepo repos.Workflow,
	dagRepo repos.DAG,
	operatorReader operator.Reader,
	artifactRepo repos.Artifact,
	dagEdgeRepo repos.DAGEdge,
	dagResultRepo repos.DAGResult,
	operatorResultRepo repos.OperatorResult,
	artifactResultRepo repos.ArtifactResult,
	vault vault.Vault,
	DB database.Database,
) error {
	// Read each workflow dag from the database that needs to be synced
	dags := make([]models.DAG, 0, len(dagIDs))
	for _, dagID := range dagIDs {
		dbDag, err := utils.ReadDAGFromDatabase(
			ctx,
			dagID,
			workflowRepo,
			dagRepo,
			operatorReader,
			artifactRepo,
			dagEdgeRepo,
			DB,
		)
		if err != nil {
			return err
		}

		dags = append(dags, *dbDag)
	}

	for _, dag := range dags {
		if err := syncWorkflowDag(
			ctx,
			&dag,
			dagRepo,
			dagResultRepo,
			operatorResultRepo,
			artifactResultRepo,
			vault,
			DB,
		); err != nil {
			log.Errorf("Unable to sync with Airflow for WorkflowDag %v: %v", dag.ID, err)
		}
	}

	return nil
}

// syncWorkflowDag fetches the latest runs from Airflow for the workflow dag
// specified and populates the database with the results.
// It returns an error, if any.
func syncWorkflowDag(
	ctx context.Context,
	dag *models.DAG,
	dagRepo repos.DAG,
	dagResultRepo repos.DAGResult,
	operatorResultRepo repos.OperatorResult,
	artifactResultRepo repos.ArtifactResult,
	vault vault.Vault,
	DB database.Database,
) error {
	// Read Airflow credentials from vault
	authConf, err := auth.ReadConfigFromSecret(
		ctx,
		dag.EngineConfig.AirflowConfig.IntegrationId,
		vault,
	)
	if err != nil {
		return err
	}

	// Create Airflow API client
	cli, err := newClient(ctx, authConf)
	if err != nil {
		return err
	}

	dagsMatch, err := checkForDAGMatch(
		ctx,
		cli,
		dag,
		dagRepo,
		DB,
	)
	if err != nil {
		return err
	}

	if !dagsMatch {
		// Skip syncing if the dags do not match
		return errors.New("The Airflow DAG does not match the Aqueduct DAG, so the workflow dag cannot be synced.")
	}

	// Get all Airflow DAG runs for `dag`
	dagRuns, err := cli.getDagRuns(dag.EngineConfig.AirflowConfig.DagId)
	if err != nil {
		return err
	}

	dagResults, err := dagResultRepo.GetByWorkflow(ctx, dag.WorkflowID, DB)
	if err != nil {
		return err
	}

	dagCreatedAtTimes := make([]time.Time, 0, len(dagResults))
	for _, dagResult := range dagResults {
		dagCreatedAtTimes = append(dagCreatedAtTimes, dagResult.CreatedAt)
	}

	for _, dagRun := range dagRuns {
		// TODO: What if this dagRun corresponds to a previous WorkflowDag?

		// Check if this DagRun has already been synced.
		// We reasonably assume that no 2 Airflow DagRuns have the same start date, because
		// the DagRun start date is measured in nanoseconds.
		if ok := timeInSlice(dagRun.GetStartDate(), dagCreatedAtTimes); ok {
			// A DagRun with the same start time has already been registered, so skip this
			continue
		}

		if *dagRun.State != airflow.DAGSTATE_SUCCESS &&
			*dagRun.State != airflow.DAGSTATE_FAILED {
			// DagRun is in either DAGSTATE_QUEUED or DAGSTATE_RUNNING,
			// i.e. it has not finished yet, so skip it.
			continue
		}

		// Populate database with WorkflowDagResult for this DagRun
		if err := syncWorkflowDagResult(
			ctx,
			cli,
			dag,
			&dagRun,
			dagResultRepo,
			operatorResultRepo,
			artifactResultRepo,
			DB,
		); err != nil {
			return err
		}
	}

	return nil
}

// syncWorkflowDagResult populates the database with a DAGResult and related
// OperatorResult(s) and ArtifactResult(s) for the Airflow DagRun `run` of the
// DAG dag. It returns an error, if any.
func syncWorkflowDagResult(
	ctx context.Context,
	cli *client,
	dag *models.DAG,
	run *airflow.DAGRun,
	dagResultRepo repos.DAGResult,
	operatorResultRepo repos.OperatorResult,
	artifactResultRepo repos.ArtifactResult,
	DB database.Database,
) error {
	txn, err := DB.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	dagResult, err := createDAGResult(
		ctx,
		dag,
		run,
		dagResultRepo,
		txn,
	)
	if err != nil {
		return err
	}

	// Get Airflow task states
	taskToState, err := cli.getTaskStates(run.GetDagId(), run.GetDagRunId())
	if err != nil {
		return err
	}

	for _, op := range dag.Operators {
		// Map Airflow task state to operator execution status
		taskID, ok := dag.EngineConfig.AirflowConfig.OperatorToTask[op.Id]
		if !ok {
			return errors.Newf("Unable to determine Airflow task ID for operator %v", op.Id)
		}

		taskState, ok := taskToState[taskID]
		if !ok {
			return errors.Newf("Unable to find Airflow task state for task %s", taskID)
		}

		execStatus := mapTaskStateToStatus(taskState)

		if err := createOperatorResult(
			ctx,
			run.GetDagRunId(),
			dag,
			&op,
			execStatus,
			dagResult.ID,
			operatorResultRepo,
			artifactResultRepo,
			txn,
		); err != nil {
			return err
		}
	}

	if err := txn.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// checkForDAGMatch checks if the Aqueduct workflow DAG dag
// matches the DAG currently registered with Airflow. They may not match if the user
// updated the workflow and has not yet copied over the updated Airflow DAG file to
// their Airflow server. If the DAGs match, it also updates dag's engine config
// in the database.
// It returns a bool whether the DAGs match and an error, if any.
func checkForDAGMatch(
	ctx context.Context,
	cli *client,
	dag *models.DAG,
	dagRepo repos.DAG,
	DB database.Database,
) (bool, error) {
	if dag.EngineConfig.AirflowConfig.MatchesAirflow {
		// We previously confirmed that the DAGs match
		return true, nil
	}

	airflowDag, err := cli.getDag(dag.EngineConfig.AirflowConfig.DagId)
	if err != nil {
		return false, err
	}

	// The way we check if the DAGs match is if dag.ID is one of tags
	// for `airflowDag`, since the workflow dag ID is set as a tag each time
	// the Airflow DAG file is generated.
	for _, tag := range airflowDag.Tags {
		if tag.GetName() == dag.ID.String() {
			// The DAGs match so the engine config needs to be updated
			dag.EngineConfig.AirflowConfig.MatchesAirflow = true
			_, err = dagRepo.Update(
				ctx,
				dag.ID,
				map[string]interface{}{
					models.DagEngineConfig: &dag.EngineConfig,
				},
				DB,
			)
			if err != nil {
				return true, err
			}

			return true, nil
		}
	}

	return false, nil
}

// timeInSlice returns whether `t` is equal to any of the elements in `s`
func timeInSlice(t time.Time, s []time.Time) bool {
	for _, tt := range s {
		if t.Equal(tt) {
			return true
		}
	}
	return false
}
