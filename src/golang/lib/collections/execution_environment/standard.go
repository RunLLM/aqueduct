package execution_environment

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type standardReaderImpl struct{}

type standardWriterImpl struct{}

func (w *standardWriterImpl) CreateExecutionEnvironment(
	ctx context.Context,
	spec *Spec,
	hash uuid.UUID,
	db database.Database,
) (*DBExecutionEnvironment, error) {
	insertColumns := []string{SpecColumn, HashColumn}
	insertStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	args := []interface{}{spec, hash}

	var executionEnvironment DBExecutionEnvironment
	err := db.Query(ctx, &executionEnvironment, insertStmt, args...)
	return &executionEnvironment, err
}

func (r *standardReaderImpl) GetExecutionEnvironment(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) (*DBExecutionEnvironment, error) {
	results, err := r.GetExecutionEnvironments(ctx, []uuid.UUID{id}, db)
	if err != nil {
		return nil, err
	}

	if len(results) != 1 {
		return nil, errors.Newf("Expected 1 result, but got %d results.", len(results))
	}

	return &results[0], nil
}

func (r *standardReaderImpl) GetExecutionEnvironments(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) ([]DBExecutionEnvironment, error) {
	if len(ids) == 0 {
		return nil, errors.New("Provided empty IDs list.")
	}

	query := fmt.Sprintf(
		"SELECT %s FROM execution_environment WHERE id IN (%s);",
		allColumns(),
		stmt_preparers.GenerateArgsList(len(ids), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(ids)

	var results []DBExecutionEnvironment
	err := db.Query(ctx, &results, query, args...)
	return results, err
}

func (r *standardReaderImpl) GetActiveExecutionEnvironmentByHash(
	ctx context.Context,
	hash uuid.UUID,
	db database.Database,
) (*DBExecutionEnvironment, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM execution_environment WHERE hash = $1 AND garbage_collected = FALSE;",
		allColumns(),
	)
	var result DBExecutionEnvironment

	err := db.Query(ctx, &result, query, hash)
	return &result, err
}

func (r *standardReaderImpl) GetActiveExecutionEnvironmentsByOperatorID(
	ctx context.Context,
	opIDs []uuid.UUID,
	db database.Database,
) (map[uuid.UUID]DBExecutionEnvironment, error) {
	type resultRow struct {
		Id               uuid.UUID `db:"id"`
		OperatorId       uuid.UUID `db:"operator_id"`
		Hash             uuid.UUID `db:"hash"`
		Spec             Spec      `db:"spec"`
		GarbageCollected bool      `db:"garbage_collected"`
	}

	query := fmt.Sprintf(`
		SELECT operator.id AS operator_id, %s
		FROM execution_environment, operator
		WHERE operator.execution_environment_id = execution_environment.id
		AND operator.id IN (%s)
		AND execution_environment.garbage_collected = FALSE;`,
		allColumnsWithPrefix(),
		stmt_preparers.GenerateArgsList(len(opIDs), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(opIDs)
	var results []resultRow
	err := db.Query(ctx, &results, query, args...)
	if err != nil {
		return nil, err
	}

	resultMap := make(map[uuid.UUID]DBExecutionEnvironment, len(results))
	for _, row := range results {
		resultMap[row.OperatorId] = DBExecutionEnvironment{
			Id:   row.Id,
			Spec: row.Spec,
			Hash: row.Hash,
		}
	}

	return resultMap, nil
}

func (w *standardWriterImpl) UpdateExecutionEnvironment(
	ctx context.Context,
	id uuid.UUID,
	changes map[string]interface{},
	db database.Database,
) (*DBExecutionEnvironment, error) {
	var result DBExecutionEnvironment
	err := utils.UpdateRecordToDest(ctx, &result, changes, tableName, IdColumn, id, allColumns(), db)
	return &result, err
}

func (w *standardWriterImpl) DeleteExecutionEnvironment(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) error {
	return w.DeleteExecutionEnvironments(ctx, []uuid.UUID{id}, db)
}

func (w *standardWriterImpl) DeleteExecutionEnvironments(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) error {
	if len(ids) == 0 {
		return nil
	}

	deleteStmt := fmt.Sprintf(
		"DELETE FROM execution_environment WHERE id IN (%s);",
		stmt_preparers.GenerateArgsList(len(ids), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(ids)
	return db.Execute(ctx, deleteStmt, args...)
}

func (r *standardReaderImpl) GetUnusedExecutionEnvironments(
	ctx context.Context,
	db database.Database,
) ([]DBExecutionEnvironment, error) {
	// Note that we use `OperatorToArtifactType` as the filtering condition because an operator
	// is guaranteed to generate at least one artifact, so this filter is guaranteed to capture
	// all operators involved in a workflow DAG.
	query := fmt.Sprintf(`
	WITH latest_workflow_dag AS
	(
		SELECT 
			workflow_dag.id 
		FROM
			workflow_dag 
		WHERE 
			created_at IN (
				SELECT 
					MAX(workflow_dag.created_at) 
				FROM 
					workflow, workflow_dag 
				WHERE 
					workflow.id = workflow_dag.workflow_id 
				GROUP BY 
					workflow.id
			)
	),
	active_execution_environment AS
	(
		SELECT DISTINCT
			operator.execution_environment_id AS id
		FROM 
			latest_workflow_dag, workflow_dag_edge, operator
		WHERE
			latest_workflow_dag.id = workflow_dag_edge.workflow_dag_id 
			AND 
			workflow_dag_edge.type = '%s' 
			AND 
			workflow_dag_edge.from_id = operator.id
	)
	SELECT 
		%s
	FROM 
		execution_environment LEFT JOIN active_execution_environment 
		ON execution_environment.id = active_execution_environment.id
	WHERE 
		execution_environment.garbage_collected = FALSE 
		AND 
		active_execution_environment.id IS NULL;`, workflow_dag_edge.OperatorToArtifactType, allColumnsWithPrefix())
	var results []DBExecutionEnvironment

	err := db.Query(ctx, &results, query)
	return results, err
}
