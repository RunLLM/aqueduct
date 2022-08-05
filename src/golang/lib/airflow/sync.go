package airflow

import (
	"context"

	"github.com/apache/airflow-client-go/airflow"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/user"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_result"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// SyncWorkflowDags syncs all dags in `workflowDagIds` with any new
// Airflow dag runs since the last sync. It returns an error, if any.
func SyncWorkflowDags(
	ctx context.Context,
	workflowDagIds []uuid.UUID,
	workflowReader workflow.Reader,
	workflowDagReader workflow_dag.Reader,
	operatorReader operator.Reader,
	artifactReader artifact.Reader,
	workflowDagEdgeReader workflow_dag_edge.Reader,
	workflowDagResultWriter workflow_dag_result.Writer,
	operatorResultWriter operator_result.Writer,
	artifactResultWriter artifact_result.Writer,
	notificationWriter notification.Writer,
	userReader user.Reader,
	vault vault.Vault,
	db database.Database,
) error {
	// Read each workflow dag from the database that needs to be synced
	dbDags := make([]workflow_dag.DBWorkflowDag, 0, len(workflowDagIds))
	for _, workflowDagId := range workflowDagIds {
		dbDag, err := utils.ReadWorkflowDagFromDatabase(
			ctx,
			workflowDagId,
			workflowReader,
			workflowDagReader,
			operatorReader,
			artifactReader,
			workflowDagEdgeReader,
			db,
		)
		if err != nil {
			return err
		}

		dbDags = append(dbDags, *dbDag)
	}

	// TODO: What happens if a workflow was updated and there are runs that haven't been synced
	// yet for a previous run.

	for _, dbDag := range dbDags {
		if err := syncWorkflowDag(
			ctx,
			&dbDag,
			workflowReader,
			workflowDagReader,
			workflowDagResultWriter,
			operatorResultWriter,
			artifactResultWriter,
			notificationWriter,
			userReader,
			vault,
			db,
		); err != nil {
			log.Errorf("Unable to sync with Airflow for WorkflowDag %v", dbDag.Id)
		}
	}

	return nil
}

// syncWorkflowDag fetches the latest runs from Airflow for the workflow dag
// specified and populates the database with the results.
// It returns an error, if any.
func syncWorkflowDag(
	ctx context.Context,
	dbDag *workflow_dag.DBWorkflowDag,
	workflowReader workflow.Reader,
	workflowDagReader workflow_dag.Reader,
	workflowDagResultWriter workflow_dag_result.Writer,
	operatorResultWriter operator_result.Writer,
	artifactResultWriter artifact_result.Writer,
	notificationWriter notification.Writer,
	userReader user.Reader,
	vault vault.Vault,
	db database.Database,
) error {
	// Read Airflow credentials from vault
	authConf, err := auth.ReadConfigFromSecret(
		ctx,
		dbDag.EngineConfig.AirflowConfig.IntegrationId,
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

	// Get all Airflow DAG runs for `dag`
	// TODO: Filter based on which runs have already been synced
	dagRuns, err := cli.getDagRuns(dbDag.EngineConfig.AirflowConfig.DagId)
	if err != nil {
		return err
	}

	txn, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	for _, dagRun := range dagRuns {
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
			dbDag,
			&dagRun,
			workflowReader,
			workflowDagReader,
			workflowDagResultWriter,
			operatorResultWriter,
			artifactResultWriter,
			notificationWriter,
			userReader,
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

// syncWorkflowDagResult populates the database with a WorkflowDagResult and related
// OperatorResult(s) and ArtifactResult(s) for the Airflow DagRun `run` of the DBWorkflowDag `dbDag`.
// It returns an error, if any.
func syncWorkflowDagResult(
	ctx context.Context,
	cli *client,
	dbDag *workflow_dag.DBWorkflowDag,
	run *airflow.DAGRun,
	workflowReader workflow.Reader,
	workflowDagReader workflow_dag.Reader,
	workflowDagResultWriter workflow_dag_result.Writer,
	operatorResultWriter operator_result.Writer,
	artifactResultWriter artifact_result.Writer,
	notificationWriter notification.Writer,
	userReader user.Reader,
	db database.Database,
) error {
	workflowDagResult, err := createWorkflowDagResult(
		ctx,
		dbDag,
		run,
		workflowReader,
		workflowDagResultWriter,
		operatorResultWriter,
		artifactResultWriter,
		notificationWriter,
		userReader,
		db,
	)
	if err != nil {
		return err
	}

	for _, op := range dbDag.Operators {
		if err := createOperatorResult(
			ctx,
			run.GetDagRunId(),
			dbDag,
			&op,
			workflowDagResult.Id,
			operatorResultWriter,
			artifactResultWriter,
			db,
		); err != nil {
			return err
		}
	}

	return nil
}
