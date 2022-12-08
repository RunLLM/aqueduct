package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type dagRepo struct {
	dagReader
	dagWriter
}

type dagReader struct{}

type dagWriter struct{}

func NewDAGRepo() repos.DAG {
	return &dagRepo{
		dagReader: dagReader{},
		dagWriter: dagWriter{},
	}
}

func (*dagReader) Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.DAG, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM workflow_dag WHERE id = $1;`,
		models.DAGCols(),
	)
	args := []interface{}{ID}

	return getDAG(ctx, DB, query, args...)
}

func (*dagReader) GetBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]models.DAG, error) {
	if len(IDs) == 0 {
		return nil, errors.New("Provided empty IDs list.")
	}

	query := fmt.Sprintf(
		`SELECT %s FROM workflow_dag WHERE id IN (%s);`,
		models.DAGCols(),
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)

	return getDAGs(ctx, DB, query, args...)
}

func (*dagReader) GetByArtifactResultBatch(ctx context.Context, artifactResultIDs []uuid.UUID, DB database.Database) (map[uuid.UUID]models.DAG, error) {
	type resultRow struct {
		ID            uuid.UUID            `db:"id"`
		WorkflowID    uuid.UUID            `db:"workflow_id"`
		CreatedAt     time.Time            `db:"created_at"`
		StorageConfig shared.StorageConfig `db:"storage_config"`
		EngineConfig  shared.EngineConfig  `db:"engine_config"`
		ArtfResultID  uuid.UUID            `db:"artf_result_id"`
	}

	query := fmt.Sprintf(`
		SELECT 
			DISTINCT artifact_result.id as artf_result_id, %s
		FROM 
			workflow_dag, workflow_dag_edge, workflow_dag_result, artifact_result
		WHERE 
			workflow_dag_edge.workflow_dag_id = workflow_dag.id
			AND (
				workflow_dag_edge.from_id = artifact_result.artifact_id
				OR 
				workflow_dag_edge.to_id = artifact_result.artifact_id
			)
			AND artifact_result.id IN (%s);`,
		models.DAGCols(),
		stmt_preparers.GenerateArgsList(len(artifactResultIDs), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(artifactResultIDs)

	var results []resultRow
	err := DB.Query(ctx, &results, query, args...)
	if err != nil {
		return nil, err
	}

	resultMap := make(map[uuid.UUID]models.DAG, len(results))
	for _, row := range results {
		resultMap[row.ArtfResultID] = models.DAG{
			ID:            row.ID,
			WorkflowID:    row.WorkflowID,
			CreatedAt:     row.CreatedAt,
			StorageConfig: row.StorageConfig,
			EngineConfig:  row.EngineConfig,
		}
	}

	return resultMap, nil
}

func (*dagReader) GetByDAGResult(ctx context.Context, dagResultID uuid.UUID, DB database.Database) (*models.DAG, error) {
	query := fmt.Sprintf(`
		SELECT %s 
		FROM workflow_dag, workflow_dag_result 
		WHERE 
			workflow_dag.id = workflow_dag_result.workflow_dag_id 
			AND workflow_dag_result.id = $1;`,
		models.DAGCols(),
	)
	args := []interface{}{dagResultID}

	return getDAG(ctx, DB, query, args...)
}

func (*dagReader) GetByOperator(ctx context.Context, operatorID uuid.UUID, DB database.Database) ([]models.DAG, error) {
	// Get all unique DAGs where there is an edge to or from the Operator with operatorID
	query := fmt.Sprintf(`
		SELECT 
			DISTINCT %s 
		FROM workflow_dag, workflow_dag_edge 
		WHERE 
			workflow_dag_edge.workflow_dag_id = workflow_dag.id 
			AND 
			(
				(workflow_dag_edge.type = '%s' AND workflow_dag_edge.from_id = $1) 
				OR 
				(workflow_dag_edge.type = '%s' AND workflow_dag_edge.to_id = $1)
			);`,
		models.DAGCols(),
		workflow_dag_edge.OperatorToArtifactType,
		workflow_dag_edge.ArtifactToOperatorType,
	)
	args := []interface{}{operatorID}

	return getDAGs(ctx, DB, query, args...)
}

func (*dagReader) GetByWorkflow(ctx context.Context, workflowID uuid.UUID, DB database.Database) ([]models.DAG, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM workflow_dag WHERE workflow_id = $1;`,
		models.DAGCols(),
	)
	args := []interface{}{workflowID}

	return getDAGs(ctx, DB, query, args...)
}

func (*dagReader) GetLatestByWorkflow(ctx context.Context, workflowID uuid.UUID, DB database.Database) (*models.DAG, error) {
	query := fmt.Sprintf(
		`SELECT %s 
		FROM workflow_dag 
		WHERE workflow_id = $1 
		ORDER BY created_at DESC LIMIT 1;`,
		models.DAGCols(),
	)
	args := []interface{}{workflowID}

	return getDAG(ctx, DB, query, args...)
}

func (*dagReader) List(ctx context.Context, DB database.Database) ([]models.DAG, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM workflow_dag;`,
		models.DAGCols(),
	)

	return getDAGs(ctx, DB, query)
}

func (*dagWriter) Create(
	ctx context.Context,
	workflowID uuid.UUID,
	storageConfig *shared.StorageConfig,
	engineConfig *shared.EngineConfig,
	DB database.Database,
) (*models.DAG, error) {
	cols := []string{
		models.DagID,
		models.DagWorkflowID,
		models.DagCreatedAt,
		models.DagStorageConfig,
		models.DagEngineConfig,
	}
	query := DB.PrepareInsertWithReturnAllStmt(models.DagTable, cols, models.DAGCols())

	ID, err := utils.GenerateUniqueUUID(ctx, models.DagTable, DB)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		ID,
		workflowID,
		time.Now(),
		storageConfig,
		engineConfig,
	}

	return getDAG(ctx, DB, query, args...)
}

func (*dagWriter) Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error {
	return deleteDAGs(ctx, DB, []uuid.UUID{ID})
}

func (*dagWriter) DeleteBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) error {
	return deleteDAGs(ctx, DB, IDs)
}

func (*dagWriter) Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, DB database.Database) (*models.DAG, error) {
	var dag models.DAG
	err := utils.UpdateRecordToDest(
		ctx,
		&dag,
		changes,
		models.DagTable,
		models.DagID,
		ID,
		models.DAGCols(),
		DB,
	)
	return &dag, err
}

func getDAGs(ctx context.Context, DB database.Database, query string, args ...interface{}) ([]models.DAG, error) {
	var dags []models.DAG
	err := DB.Query(ctx, &dags, query, args...)
	return dags, err
}

func getDAG(ctx context.Context, DB database.Database, query string, args ...interface{}) (*models.DAG, error) {
	dags, err := getDAGs(ctx, DB, query, args...)
	if err != nil {
		return nil, err
	}

	if len(dags) == 0 {
		return nil, database.ErrNoRows
	}

	if len(dags) != 1 {
		return nil, errors.Newf("Expected 1 DAG but got %v", len(dags))
	}

	return &dags[0], nil
}

func deleteDAGs(ctx context.Context, DB database.Database, IDs []uuid.UUID) error {
	if len(IDs) == 0 {
		return nil
	}

	query := fmt.Sprintf(
		`DELETE FROM workflow_dag WHERE id IN (%s)`,
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)

	return DB.Execute(ctx, query, args...)
}
