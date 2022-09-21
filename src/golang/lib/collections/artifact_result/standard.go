package artifact_result

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type standardReaderImpl struct{}

type standardWriterImpl struct{}

func (w *standardWriterImpl) CreateArtifactResult(
	ctx context.Context,
	workflowDagResultId uuid.UUID,
	artifactId uuid.UUID,
	contentPath string,
	db database.Database,
) (*ArtifactResult, error) {
	insertColumns := []string{WorkflowDagResultIdColumn, ArtifactIdColumn, ContentPathColumn, StatusColumn}
	insertArtifactStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	args := []interface{}{workflowDagResultId, artifactId, contentPath, shared.PendingExecutionStatus}

	var artifactResult ArtifactResult
	err := db.Query(ctx, &artifactResult, insertArtifactStmt, args...)
	return &artifactResult, err
}

func (w *standardWriterImpl) InsertArtifactResult(
	ctx context.Context,
	workflowDagResultId uuid.UUID,
	artifactId uuid.UUID,
	contentPath string,
	execState *shared.ExecutionState,
	metadata *Metadata,
	db database.Database,
) (*ArtifactResult, error) {
	insertColumns := []string{WorkflowDagResultIdColumn, ArtifactIdColumn, ContentPathColumn, StatusColumn, ExecStateColumn, MetadataColumn}
	insertArtifactStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	args := []interface{}{workflowDagResultId, artifactId, contentPath, execState.Status, execState, metadata}

	var artifactResult ArtifactResult
	err := db.Query(ctx, &artifactResult, insertArtifactStmt, args...)
	return &artifactResult, err
}

func (r *standardReaderImpl) GetArtifactResult(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) (*ArtifactResult, error) {
	artifactResults, err := r.GetArtifactResults(ctx, []uuid.UUID{id}, db)
	if err != nil {
		return nil, err
	}

	if len(artifactResults) != 1 {
		return nil, errors.Newf("Expected 1 artifact_result, but got %d artifact_results.", len(artifactResults))
	}

	return &artifactResults[0], nil
}

func (r *standardReaderImpl) GetArtifactResults(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) ([]ArtifactResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("Provided empty IDs list.")
	}

	getArtifactResultsQuery := fmt.Sprintf(
		"SELECT %s FROM artifact_result WHERE id IN (%s);",
		allColumns(),
		stmt_preparers.GenerateArgsList(len(ids), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(ids)

	var artifactResults []ArtifactResult
	err := db.Query(ctx, &artifactResults, getArtifactResultsQuery, args...)
	return artifactResults, err
}

func (r *standardReaderImpl) GetArtifactResultsByArtifactId(
	ctx context.Context,
	artifactId uuid.UUID,
	db database.Database,
) ([]ArtifactResult, error) {
	getArtifactResultsQuery := fmt.Sprintf(
		"SELECT %s FROM artifact_result WHERE artifact_id = $1;",
		allColumns(),
	)

	var artifactResults []ArtifactResult
	err := db.Query(ctx, &artifactResults, getArtifactResultsQuery, artifactId)
	return artifactResults, err
}

func (r *standardReaderImpl) GetArtifactResultsByArtifactNameAndWorkflowId(
	ctx context.Context,
	workflowId uuid.UUID,
	name string,
	db database.Database,
) ([]ArtifactResult, error) {
	getArtifactResultsQuery := fmt.Sprintf(`
		SELECT DISTINCT %s FROM artifact_result, artifact, workflow_dag, workflow_dag_edge
		WHERE workflow_dag.workflow_id = $1
		AND artifact.name = $2
		AND (
			workflow_dag_edge.from_id = artifact.id
			OR
			workflow_dag_edge.to_id = artifact.id
		)
		AND workflow_dag_edge.workflow_dag_id = workflow_dag.id
		AND artifact_result.artifact_id = artifact.id;`,
		allColumnsWithPrefix(),
	)

	var artifactResults []ArtifactResult
	err := db.Query(ctx, &artifactResults, getArtifactResultsQuery, workflowId, name)
	return artifactResults, err
}

func (r *standardReaderImpl) GetArtifactResultByWorkflowDagResultIdAndArtifactId(
	ctx context.Context,
	workflowDagResultId, artifactId uuid.UUID,
	db database.Database,
) (*ArtifactResult, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM artifact_result WHERE workflow_dag_result_id = $1 AND artifact_id = $2;",
		allColumns(),
	)

	var artifactResult ArtifactResult
	err := db.Query(ctx, &artifactResult, query, workflowDagResultId, artifactId)
	return &artifactResult, err
}

func (r *standardReaderImpl) GetArtifactResultsByWorkflowDagResultIds(
	ctx context.Context,
	workflowDagResultIds []uuid.UUID,
	db database.Database,
) ([]ArtifactResult, error) {
	if len(workflowDagResultIds) == 0 {
		return nil, errors.New("Provided empty IDs list.")
	}

	query := fmt.Sprintf(
		"SELECT %s FROM artifact_result WHERE workflow_dag_result_id IN (%s);",
		allColumns(),
		stmt_preparers.GenerateArgsList(len(workflowDagResultIds), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(workflowDagResultIds)

	var artifactResults []ArtifactResult
	err := db.Query(ctx, &artifactResults, query, args...)
	return artifactResults, err
}

func (w *standardWriterImpl) UpdateArtifactResult(
	ctx context.Context,
	id uuid.UUID,
	changes map[string]interface{},
	db database.Database,
) (*ArtifactResult, error) {
	var artifactResult ArtifactResult
	err := utils.UpdateRecordToDest(ctx, &artifactResult, changes, tableName, IdColumn, id, allColumns(), db)
	return &artifactResult, err
}

func (w *standardWriterImpl) DeleteArtifactResult(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) error {
	return w.DeleteArtifactResults(ctx, []uuid.UUID{id}, db)
}

func (w *standardWriterImpl) DeleteArtifactResults(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) error {
	if len(ids) == 0 {
		return nil
	}

	deleteArtifactResultStmt := fmt.Sprintf(
		"DELETE FROM artifact_result WHERE id IN (%s);",
		stmt_preparers.GenerateArgsList(len(ids), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(ids)
	return db.Execute(ctx, deleteArtifactResultStmt, args...)
}
