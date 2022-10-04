package workflow_dag

import (
	"context"
	"fmt"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type standardReaderImpl struct{}

type standardWriterImpl struct{}

func (w *standardWriterImpl) CreateWorkflowDag(
	ctx context.Context,
	workflowId uuid.UUID,
	storageConfig *shared.StorageConfig,
	engineConfig *shared.EngineConfig,
	db database.Database,
) (*DBWorkflowDag, error) {
	insertColumns := []string{WorkflowIdColumn, CreatedAtColumn, StorageConfigColumn, EngineConfigColumn}
	insertWorkflowDagStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	args := []interface{}{workflowId, time.Now(), storageConfig, engineConfig}

	var workflowDag DBWorkflowDag
	err := db.Query(ctx, &workflowDag, insertWorkflowDagStmt, args...)
	return &workflowDag, err
}

func (r *standardReaderImpl) GetWorkflowDag(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) (*DBWorkflowDag, error) {
	workflowDags, err := r.GetWorkflowDags(ctx, []uuid.UUID{id}, db)
	if err != nil {
		return nil, err
	}

	if len(workflowDags) != 1 {
		return nil, errors.Newf("Expected 1 workflow_dag, but got %d workflow_dags.", len(workflowDags))
	}

	return &workflowDags[0], nil
}

func (r *standardReaderImpl) GetWorkflowDags(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) ([]DBWorkflowDag, error) {
	if len(ids) == 0 {
		return nil, errors.New("Provided empty IDs list.")
	}

	getWorkflowDagsQuery := fmt.Sprintf(
		"SELECT %s FROM workflow_dag WHERE id IN (%s);",
		allColumns(),
		stmt_preparers.GenerateArgsList(len(ids), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(ids)

	var workflowDags []DBWorkflowDag
	err := db.Query(ctx, &workflowDags, getWorkflowDagsQuery, args...)
	return workflowDags, err
}

func (r *standardReaderImpl) GetWorkflowDagsByWorkflowId(
	ctx context.Context,
	workflowId uuid.UUID,
	db database.Database,
) ([]DBWorkflowDag, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM workflow_dag WHERE workflow_id = $1;",
		allColumns(),
	)

	var workflowDags []DBWorkflowDag
	err := db.Query(ctx, &workflowDags, query, workflowId)
	return workflowDags, err
}

func (r *standardReaderImpl) GetLatestWorkflowDag(
	ctx context.Context,
	workflowId uuid.UUID,
	db database.Database,
) (*DBWorkflowDag, error) {
	getLatestWorkflowDagQuery := fmt.Sprintf(
		"SELECT %s FROM workflow_dag WHERE workflow_id = $1 ORDER BY created_at DESC LIMIT 1;",
		allColumns(),
	)

	var workflowDag DBWorkflowDag
	err := db.Query(ctx, &workflowDag, getLatestWorkflowDagQuery, workflowId)
	return &workflowDag, err
}

func (r *standardReaderImpl) GetWorkflowDagByWorkflowDagResultId(
	ctx context.Context,
	workflowDagResultId uuid.UUID,
	db database.Database,
) (*DBWorkflowDag, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM workflow_dag, workflow_dag_result 
		WHERE workflow_dag.id = workflow_dag_result.workflow_dag_id 
		AND workflow_dag_result.id = $1;`,
		allColumnsWithPrefix(),
	)

	var workflowDag DBWorkflowDag
	err := db.Query(ctx, &workflowDag, query, workflowDagResultId)
	return &workflowDag, err
}

func (r *standardReaderImpl) GetWorkflowDagsByOperatorId(
	ctx context.Context,
	operatorId uuid.UUID,
	db database.Database,
) ([]DBWorkflowDag, error) {
	// Get all unique workflow DAGs which has an edge to the operator node with the id `operatorId` (is `from_id` or `to_id`)
	query := fmt.Sprintf(`
		SELECT DISTINCT %s FROM workflow_dag, workflow_dag_edge 
		WHERE workflow_dag_edge.workflow_dag_id = workflow_dag.id AND 
		((workflow_dag_edge.type = '%s' AND workflow_dag_edge.from_id = $1) OR 
		(workflow_dag_edge.type = '%s' AND workflow_dag_edge.to_id = $1));`,
		allColumnsWithPrefix(), workflow_dag_edge.OperatorToArtifactType, workflow_dag_edge.ArtifactToOperatorType)

	var workflowDags []DBWorkflowDag
	err := db.Query(ctx, &workflowDags, query, operatorId)
	return workflowDags, err
}

func (r *standardReaderImpl) GetWorkflowDagsMapByArtifactResultIds(
	ctx context.Context,
	artifactResultIds []uuid.UUID,
	db database.Database,
) (map[uuid.UUID]DBWorkflowDag, error) {
	type resultRow struct {
		Id            uuid.UUID            `db:"id"`
		WorkflowId    uuid.UUID            `db:"workflow_id"`
		CreatedAt     time.Time            `db:"created_at"`
		StorageConfig shared.StorageConfig `db:"storage_config"`
		EngineConfig  shared.EngineConfig  `db:"engine_config"`
		ArtfResultId  uuid.UUID            `db:"artf_result_id"`
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT artifact_result.id as artf_result_id, %s
		FROM workflow_dag, workflow_dag_edge, workflow_dag_result, artifact_result
		WHERE workflow_dag_edge.workflow_dag_id = workflow_dag.id
		AND (
			workflow_dag_edge.from_id = artifact_result.artifact_id
		OR 
			workflow_dag_edge.to_id = artifact_result.artifact_id
		)
		AND artifact_result.id IN (%s);`,
		allColumnsWithPrefix(),
		stmt_preparers.GenerateArgsList(len(artifactResultIds), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(artifactResultIds)
	var results []resultRow
	err := db.Query(ctx, &results, query, args...)
	if err != nil {
		return nil, err
	}

	resultMap := make(map[uuid.UUID]DBWorkflowDag, len(results))
	for _, row := range results {
		resultMap[row.ArtfResultId] = DBWorkflowDag{
			Id:            row.Id,
			WorkflowId:    row.WorkflowId,
			CreatedAt:     row.CreatedAt,
			StorageConfig: row.StorageConfig,
			EngineConfig:  row.EngineConfig,
		}
	}

	return resultMap, nil
}

func (w *standardWriterImpl) UpdateWorkflowDag(
	ctx context.Context,
	id uuid.UUID,
	changes map[string]interface{},
	db database.Database,
) (*DBWorkflowDag, error) {
	var workflowDag DBWorkflowDag
	err := utils.UpdateRecordToDest(ctx, &workflowDag, changes, tableName, IdColumn, id, allColumns(), db)
	return &workflowDag, err
}

func (w *standardWriterImpl) DeleteWorkflowDag(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) error {
	return w.DeleteWorkflowDags(ctx, []uuid.UUID{id}, db)
}

func (w *standardWriterImpl) DeleteWorkflowDags(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) error {
	if len(ids) == 0 {
		return nil
	}

	deleteWorkflowDagsStmt := fmt.Sprintf(
		"DELETE FROM workflow_dag WHERE id IN (%s);",
		stmt_preparers.GenerateArgsList(len(ids), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(ids)
	return db.Execute(ctx, deleteWorkflowDagsStmt, args...)
}
