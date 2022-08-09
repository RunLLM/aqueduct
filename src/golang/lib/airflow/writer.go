package airflow

import (
	"context"
	"fmt"

	"github.com/apache/airflow-client-go/airflow"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/user"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_result"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func createWorkflowDagResult(
	ctx context.Context,
	dbDag *workflow_dag.DBWorkflowDag,
	run *airflow.DAGRun,
	workflowReader workflow.Reader,
	workflowDagResultWriter workflow_dag_result.Writer,
	notificationWriter notification.Writer,
	userReader user.Reader,
	db database.Database,
) (*workflow_dag_result.WorkflowDagResult, error) {
	workflowDagStatus := mapDagStateToStatus(*run.State)
	if workflowDagStatus != shared.SucceededExecutionStatus &&
		workflowDagStatus != shared.FailedExecutionStatus {
		// Do not create WorkflowDagResult for Airflow DAG runs that have not finished
		return nil, errors.New("Cannot create WorkflowDagResult for in progress Airflow DAG Run.")
	}

	workflowDagResult, err := workflowDagResultWriter.CreateWorkflowDagResult(ctx, dbDag.Id, db)
	if err != nil {
		return nil, err
	}

	changes := map[string]interface{}{
		workflow_dag_result.CreatedAtColumn: *run.StartDate.Get(),
		workflow_dag_result.StatusColumn:    workflowDagStatus,
	}

	return workflowDagResultWriter.UpdateWorkflowDagResult(
		ctx,
		workflowDagResult.Id,
		changes,
		workflowReader,
		notificationWriter,
		userReader,
		db,
	)
}

func createOperatorResult(
	ctx context.Context,
	dagRunId string,
	dbDag *workflow_dag.DBWorkflowDag,
	dbOp *operator.DBOperator,
	execStatus shared.ExecutionStatus,
	workflowDagResultId uuid.UUID,
	operatorResultWriter operator_result.Writer,
	artifactResultWriter artifact_result.Writer,
	db database.Database,
) error {
	logrus.Warnf("Creating Result for Operator ID: %v", dbOp.Id)

	// Read Operator metadata to determine ExecutionState
	metadataPathPrefix, ok := dbDag.EngineConfig.AirflowConfig.OperatorMetadataPathPrefix[dbOp.Id]
	if !ok {
		return errors.Newf("Unable to find metadata path for operator %v", dbOp.Id)
	}
	metadataPath := getOperatorMetadataPath(metadataPathPrefix, dagRunId)

	execState := getOperatorExecState(ctx, execStatus, &dbDag.StorageConfig, metadataPath)

	logrus.Warnf("Using op exec state: %v", *execState)

	// Insert OperatorResult
	_, err := operatorResultWriter.InsertOperatorResult(
		ctx,
		workflowDagResultId,
		dbOp.Id,
		execState,
		db,
	)
	if err != nil {
		return err
	}

	// Insert an ArtifactResults for each output artifact
	for _, artifactId := range dbOp.Outputs {
		if err := createArtifactResult(
			ctx,
			dagRunId,
			dbDag,
			workflowDagResultId,
			artifactId,
			execState,
			artifactResultWriter,
			db,
		); err != nil {
			return err
		}
	}

	return nil
}

func createArtifactResult(
	ctx context.Context,
	dagRunId string,
	dbDag *workflow_dag.DBWorkflowDag,
	workflowDagResultId uuid.UUID,
	artifactId uuid.UUID,
	execState *shared.ExecutionState,
	artifactResultWriter artifact_result.Writer,
	db database.Database,
) error {
	// Read Artifact metadata
	metadataPathPrefix, ok := dbDag.EngineConfig.AirflowConfig.ArtifactMetadataPathPrefix[artifactId]
	if !ok {
		return errors.Newf("Unable to find metadata path for artifact %v", artifactId)
	}
	metadataPath := getArtifactMetadataPath(metadataPathPrefix, dagRunId)

	var metadata artifact_result.Metadata
	if err := utils.ReadFromStorage(
		ctx,
		&dbDag.StorageConfig,
		metadataPath,
		&metadata,
	); err != nil {
		return err
	}

	contentPathPrefix, ok := dbDag.EngineConfig.AirflowConfig.ArtifactContentPathPrefix[artifactId]
	if !ok {
		return errors.Newf("Unable to find content path for artifact %v", artifactId)
	}
	contentPath := getArtifactContentPath(contentPathPrefix, dagRunId)

	_, err := artifactResultWriter.InsertArtifactResult(
		ctx,
		workflowDagResultId,
		artifactId,
		contentPath,
		execState,
		&metadata,
		db,
	)

	return err
}

func getOperatorExecState(
	ctx context.Context,
	execStatus shared.ExecutionStatus,
	storageConfig *shared.StorageConfig,
	metadataPath string,
) *shared.ExecutionState {
	logrus.Warnf("Trying to read operator metadata from: %v", metadataPath)
	if execStatus == shared.PendingExecutionStatus {
		return &shared.ExecutionState{
			Status: shared.PendingExecutionStatus,
		}
	}

	if !utils.ObjectExistsInStorage(ctx, storageConfig, metadataPath) {
		// Metadata does not exist, so just use the state determined via the Airflow TaskState
		return &shared.ExecutionState{
			Status: execStatus,
		}
	}

	var execState shared.ExecutionState
	err := utils.ReadFromStorage(
		ctx,
		storageConfig,
		metadataPath,
		&execState,
	)
	if err != nil {
		failureType := shared.SystemFailure
		return &shared.ExecutionState{
			Status:      shared.FailedExecutionStatus,
			FailureType: &failureType,
			Error: &shared.Error{
				Context: fmt.Sprintf("%v", err),
				Tip:     shared.TipUnknownInternalError,
			},
		}
	}

	return &execState
}
