package airflow

import (
	"context"
	"fmt"

	"github.com/apache/airflow-client-go/airflow"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	mdl_shared "github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

func createDAGResult(
	ctx context.Context,
	dag *models.DAG,
	run *airflow.DAGRun,
	dagResultRepo repos.DAGResult,
	DB database.Database,
) (*models.DAGResult, error) {
	dagStatus := mapDagStateToStatus(*run.State)
	if dagStatus != mdl_shared.SucceededExecutionStatus &&
		dagStatus != mdl_shared.FailedExecutionStatus {
		// Do not create WorkflowDagResult for Airflow DAG runs that have not finished
		return nil, errors.New("Cannot create WorkflowDagResult for in progress Airflow DAG Run.")
	}

	return dagResultRepo.Create(
		ctx,
		dag.ID,
		&mdl_shared.ExecutionState{
			Status: dagStatus,
			Timestamps: &mdl_shared.ExecutionTimestamps{
				PendingAt:  run.StartDate.Get(),
				RunningAt:  run.StartDate.Get(),
				FinishedAt: run.EndDate.Get(),
			},
		},
		DB,
	)
}

func createOperatorResult(
	ctx context.Context,
	dagRunId string,
	dag *models.DAG,
	dbOp *models.Operator,
	execStatus shared.ExecutionStatus,
	dagResultID uuid.UUID,
	operatorResultRepo repos.OperatorResult,
	artifactResultRepo repos.ArtifactResult,
	DB database.Database,
) error {
	// Read Operator metadata to determine ExecutionState
	metadataPathPrefix, ok := dag.EngineConfig.AirflowConfig.OperatorMetadataPathPrefix[dbOp.ID]
	if !ok {
		return errors.Newf("Unable to find metadata path for operator %v", dbOp.ID)
	}
	metadataPath := getOperatorMetadataPath(metadataPathPrefix, dagRunId)

	// Use combination of the Airflow Task State and operator metadata to determine execution state
	execState := getOperatorExecState(ctx, execStatus, &dag.StorageConfig, metadataPath)

	// Insert OperatorResult
	_, err := operatorResultRepo.Create(
		ctx,
		dagResultID,
		dbOp.ID,
		execState,
		DB,
	)
	if err != nil {
		return err
	}

	// Insert an ArtifactResults for each output artifact
	for _, artifactId := range dbOp.Outputs {
		if err := createArtifactResult(
			ctx,
			dagRunId,
			dag,
			dagResultID,
			artifactId,
			execState,
			artifactResultRepo,
			DB,
		); err != nil {
			return err
		}
	}

	return nil
}

func createArtifactResult(
	ctx context.Context,
	dagRunId string,
	dag *models.DAG,
	dagResultID uuid.UUID,
	artifactID uuid.UUID,
	execState *shared.ExecutionState,
	artifactResultRepo repos.ArtifactResult,
	DB database.Database,
) error {
	// Read Artifact metadata
	metadataPathPrefix, ok := dag.EngineConfig.AirflowConfig.ArtifactMetadataPathPrefix[artifactID]
	if !ok {
		return errors.Newf("Unable to find metadata path for artifact %v", artifactID)
	}
	metadataPath := getArtifactMetadataPath(metadataPathPrefix, dagRunId)

	var metadata artifact_result.Metadata
	if utils.ObjectExistsInStorage(ctx, &dag.StorageConfig, metadataPath) {
		if err := utils.ReadFromStorage(
			ctx,
			&dag.StorageConfig,
			metadataPath,
			&metadata,
		); err != nil {
			return err
		}
	}

	contentPathPrefix, ok := dag.EngineConfig.AirflowConfig.ArtifactContentPathPrefix[artifactID]
	if !ok {
		return errors.Newf("Unable to find content path for artifact %v", artifactID)
	}
	contentPath := getArtifactContentPath(contentPathPrefix, dagRunId)

	_, err := artifactResultRepo.CreateWithExecStateAndMetadata(
		ctx,
		dagResultID,
		artifactID,
		contentPath,
		execState,
		&metadata,
		DB,
	)

	return err
}

func getOperatorExecState(
	ctx context.Context,
	execStatus shared.ExecutionStatus,
	storageConfig *shared.StorageConfig,
	metadataPath string,
) *shared.ExecutionState {
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
