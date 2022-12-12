package sqlite

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/aqueducthq/aqueduct/lib/models"
	mdl_shared "github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type artifactResultRepo struct {
	artifactResultReader
	artifactResultWriter
}

type artifactResultReader struct{}

type artifactResultWriter struct{}

func NewArtifactResultRepo() repos.ArtifactResult {
	return &artifactResultRepo{
		artifactResultReader: artifactResultReader{},
		artifactResultWriter: artifactResultWriter{},
	}
}

func (*artifactResultReader) Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.ArtifactResult, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM artifact_result WHERE id = $1`,
		models.ArtifactResultCols(),
	)
	args := []interface{}{ID}

	return getArtifactResult(ctx, DB, query, args...)
}

func (*artifactResultReader) GetBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]models.ArtifactResult, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM artifact_result WHERE id IN (%s);`,
		models.ArtifactResultCols(),
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)

	return getArtifactResults(ctx, DB, query, args...)
}

func (*artifactResultReader) GetByArtifact(ctx context.Context, artifactID uuid.UUID, DB database.Database) ([]models.ArtifactResult, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM artifact_result WHERE artifact_id = $1;`,
		models.ArtifactResultCols(),
	)
	args := []interface{}{artifactID}
	return getArtifactResults(ctx, DB, query, args...)
}

func (*artifactResultReader) GetByArtifactNameAndWorkflow(
	ctx context.Context,
	artifactName string,
	workflowID uuid.UUID,
	DB database.Database,
) ([]models.ArtifactResult, error) {
	query := fmt.Sprintf(
		`SELECT DISTINCT %s FROM artifact_result, artifact, workflow_dag, workflow_dag_edge
		WHERE workflow_dag.workflow_id = $1
		AND artifact.name = $2
		AND (
			workflow_dag_edge.from_id = artifact.id
			OR
			workflow_dag_edge.to_id = artifact.id
		)
		AND workflow_dag_edge.workflow_dag_id = workflow_dag.id
		AND artifact_result.artifact_id = artifact.id;`,
		models.ArtifactResultColsWithPrefix(),
	)
	args := []interface{}{workflowID, artifactName}
	return getArtifactResults(ctx, DB, query, args...)
}

func (*artifactResultReader) GetByArtifactAndDAGResult(
	ctx context.Context,
	artifactID uuid.UUID,
	dagResultID uuid.UUID,
	DB database.Database,
) (*models.ArtifactResult, error) {
	query := fmt.Sprintf(
		`SELECT %s 
		FROM artifact_result 
		WHERE workflow_dag_result_id = $1 AND artifact_id = $2;`,
		models.ArtifactResultCols(),
	)
	args := []interface{}{dagResultID, artifactID}
	return getArtifactResult(ctx, DB, query, args...)
}

func (*artifactResultReader) GetByDAGResults(ctx context.Context, dagResultIDs []uuid.UUID, DB database.Database) ([]models.ArtifactResult, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM artifact_result WHERE workflow_dag_result_id IN (%s);`,
		models.ArtifactResultColsWithPrefix(),
		stmt_preparers.GenerateArgsList(len(dagResultIDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(dagResultIDs)

	return getArtifactResults(ctx, DB, query, args...)
}

func (*artifactResultReader) GetStatusByArtifactBatch(
	ctx context.Context,
	artifactIDs []uuid.UUID,
	DB database.Database,
) ([]views.ArtifactResultStatus, error) {
	query := fmt.Sprintf(
		`SELECT 
			artifact_result.artifact_id, 
			artifact_result.id as artifact_result_id,
			artifact_result.workflow_dag_result_id, 
			artifact_result.status,
			artifact_result.content_path,
			artifact_result.metadata,
			workflow_dag_result.created_at AS timestamp 
		FROM artifact_result, workflow_dag_result 
		WHERE 
			artifact_result.workflow_dag_result_id = workflow_dag_result.id 
			AND artifact_result.artifact_id IN (%s);`,
		stmt_preparers.GenerateArgsList(len(artifactIDs), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(artifactIDs)

	var statuses []views.ArtifactResultStatus
	err := DB.Query(ctx, &statuses, query, args...)
	return statuses, err
}

func (*artifactResultReader) GetByArtifactBatch(
	ctx context.Context,
	artifactIDs []uuid.UUID,
	DB database.Database,
) ([]models.ArtifactResult, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM artifact_result
		WHERE artifact_result.artifact_id IN (%s);`,
		models.ArtifactResultCols(),
		stmt_preparers.GenerateArgsList(len(artifactIDs), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(artifactIDs)

	var results []models.ArtifactResult
	err := DB.Query(ctx, &results, query, args...)
	return results, err
}

func (*artifactResultReader) GetWithArtifactOfMetricsByDAGResultBatch(
	ctx context.Context,
	dagResultIDs []uuid.UUID,
	DB database.Database,
) ([]views.ArtifactWithResult, error) {
	query := fmt.Sprintf(
		`SELECT DISTINCT
			artifact.id as id,
			artifact.name as name,
			artifact.description as description,
			artifact.type as type,
			artifact_result.id as result_id,
			artifact_result.workflow_dag_result_id as dag_result_id,
			artifact_result.content_path as content_path,
			artifact_result.execution_state as execution_state,
			artifact_result.metadata as metadata,
			workflow_dag.storage_config as storage_config
		FROM
			workflow_dag,
			workflow_dag_edge,
			operator,
			artifact,
			artifact_result
		WHERE 
			workflow_dag_edge.to_id = artifact.id
			AND workflow_dag_edge.from_id = operator.id
			AND workflow_dag_edge.workflow_dag_id = workflow_dag.id
			AND json_extract(operator.spec, '$.type') = '%s'
			AND artifact_result.artifact_id = artifact.id
			AND artifact_result.workflow_dag_result_id IN (%s);`,
		mdl_shared.MetricType,
		stmt_preparers.GenerateArgsList(len(dagResultIDs), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(dagResultIDs)
	var results []views.ArtifactWithResult

	err := DB.Query(ctx, &results, query, args...)
	return results, err
}

func (*artifactResultWriter) Create(
	ctx context.Context,
	dagResultID uuid.UUID,
	artifactID uuid.UUID,
	contentPath string,
	DB database.Database,
) (*models.ArtifactResult, error) {
	cols := []string{
		models.ArtifactResultID,
		models.ArtifactResultDAGResultID,
		models.ArtifactResultArtifactID,
		models.ArtifactResultContentPath,
		models.ArtifactResultStatus,
	}
	query := DB.PrepareInsertWithReturnAllStmt(models.ArtifactResultTable, cols, models.ArtifactResultCols())

	ID, err := utils.GenerateUniqueUUID(ctx, models.ArtifactResultTable, DB)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		ID,
		dagResultID,
		artifactID,
		contentPath,
		shared.PendingExecutionStatus,
	}
	return getArtifactResult(ctx, DB, query, args...)
}

func (*artifactResultWriter) CreateWithExecStateAndMetadata(
	ctx context.Context,
	dagResultID uuid.UUID,
	artifactID uuid.UUID,
	contentPath string,
	execState *shared.ExecutionState,
	metadata *artifact_result.Metadata,
	DB database.Database,
) (*models.ArtifactResult, error) {
	cols := []string{
		models.ArtifactResultID,
		models.ArtifactResultDAGResultID,
		models.ArtifactResultArtifactID,
		models.ArtifactResultContentPath,
		models.ArtifactResultStatus,
		models.ArtifactResultMetadata,
		models.ArtifactResultExecState,
	}
	query := DB.PrepareInsertWithReturnAllStmt(models.ArtifactResultTable, cols, models.ArtifactResultCols())

	ID, err := utils.GenerateUniqueUUID(ctx, models.ArtifactResultTable, DB)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		ID,
		dagResultID,
		artifactID,
		contentPath,
		execState.Status,
		metadata,
		execState,
	}
	return getArtifactResult(ctx, DB, query, args...)
}

func (*artifactResultWriter) Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error {
	return deleteArtifactResults(ctx, DB, []uuid.UUID{ID})
}

func (*artifactResultWriter) DeleteBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) error {
	return deleteArtifactResults(ctx, DB, IDs)
}

func (*artifactResultWriter) Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, DB database.Database) (*models.ArtifactResult, error) {
	var artifactResult models.ArtifactResult
	err := utils.UpdateRecordToDest(
		ctx,
		&artifactResult,
		changes,
		models.ArtifactResultTable,
		models.ArtifactResultID,
		ID,
		models.ArtifactResultCols(),
		DB,
	)
	return &artifactResult, err
}

func deleteArtifactResults(ctx context.Context, DB database.Database, IDs []uuid.UUID) error {
	if len(IDs) == 0 {
		return nil
	}

	query := fmt.Sprintf(
		`DELETE FROM artifact_result WHERE id IN (%s)`,
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)

	return DB.Execute(ctx, query, args...)
}

func getArtifactResults(ctx context.Context, DB database.Database, query string, args ...interface{}) ([]models.ArtifactResult, error) {
	var artifactResults []models.ArtifactResult
	err := DB.Query(ctx, &artifactResults, query, args...)
	return artifactResults, err
}

func getArtifactResult(ctx context.Context, DB database.Database, query string, args ...interface{}) (*models.ArtifactResult, error) {
	artifactResults, err := getArtifactResults(ctx, DB, query, args...)
	if err != nil {
		return nil, err
	}

	if len(artifactResults) == 0 {
		return nil, database.ErrNoRows
	}

	if len(artifactResults) != 1 {
		return nil, errors.Newf("Expected 1 artifactResult but got %v", len(artifactResults))
	}

	return &artifactResults[0], nil
}
