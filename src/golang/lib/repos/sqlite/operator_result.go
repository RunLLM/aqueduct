package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type operatorResultRepo struct {
	operatorResultReader
	operatorResultWriter
}

type operatorResultReader struct{}

type operatorResultWriter struct{}

func NewOperatorResultRepo() repos.OperatorResult {
	return &operatorResultRepo{
		operatorResultReader: operatorResultReader{},
		operatorResultWriter: operatorResultWriter{},
	}
}

func (*operatorResultReader) Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.OperatorResult, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM operator_result WHERE id = $1;`,
		models.OperatorResultCols(),
	)
	args := []interface{}{ID}

	return getOperatorResult(ctx, DB, query, args...)
}

func (*operatorResultReader) GetBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]models.OperatorResult, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM operator_result WHERE id IN (%s);`,
		models.OperatorResultCols(),
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)

	return getOperatorResults(ctx, DB, query, args...)
}

func (*operatorResultReader) GetByDAGResultAndOperator(
	ctx context.Context,
	dagResultID uuid.UUID,
	operatorID uuid.UUID,
	DB database.Database,
) (*models.OperatorResult, error) {
	query := fmt.Sprintf(
		`SELECT %s
		FROM operator_result
		WHERE workflow_dag_result_id = $1 AND operator_id = $2;`,
		models.OperatorResultCols(),
	)
	args := []interface{}{dagResultID, operatorID}

	return getOperatorResult(ctx, DB, query, args...)
}

func (*operatorResultReader) GetWithOperatorByDAGResultBatch(
	ctx context.Context,
	dagResultIDs []uuid.UUID,
	types []operator.Type,
	DB database.Database,
) ([]views.OperatorWithResult, error) {
	query := fmt.Sprintf(
		`SELECT
			operator.id as id,
			operator.name as name,
			operator.description as description,
			operator.spec as spec,
			operator.execution_environment_id as execution_environment_id,
			operator_result.id as result_id,
			operator_result.workflow_dag_result_id as dag_result_id,
			operator_result.status as status,
			operator_result.execution_state as execution_state
		FROM operator, operator_result 
		WHERE operator_result.workflow_dag_result_id IN (%s)
		AND json_extract(operator.spec, '$.type') IN (%s)
		AND operator.id = operator_result.operator_id`,
		stmt_preparers.GenerateArgsList(len(dagResultIDs), 1),
		stmt_preparers.GenerateArgsList(len(types), 1+len(dagResultIDs)),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(dagResultIDs)
	for _, tp := range types {
		args = append(args, tp)
	}

	var results []views.OperatorWithResult
	err := DB.Query(ctx, &results, query, args...)
	return results, err
}

func (*operatorResultReader) GetByDAGResultBatch(
	ctx context.Context,
	dagResultIDs []uuid.UUID,
	DB database.Database,
) ([]models.OperatorResult, error) {
	query := fmt.Sprintf(
		`SELECT %s 
		FROM operator_result 
		WHERE workflow_dag_result_id IN (%s);`,
		models.OperatorResultCols(),
		stmt_preparers.GenerateArgsList(len(dagResultIDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(dagResultIDs)

	return getOperatorResults(ctx, DB, query, args...)
}

func (*operatorResultReader) GetCheckStatusByArtifactBatch(
	ctx context.Context,
	artifactIDs []uuid.UUID,
	DB database.Database,
) ([]views.OperatorResultStatus, error) {
	// Get all unique combinations of artifact id, operator name,
	// operator status, operator execution state, and workflow dag
	// result id of all check operators of artifacts in the
	// `artifactIds` list (`from_id` in `artifactIds`).
	query := fmt.Sprintf(
		`SELECT DISTINCT
			workflow_dag_edge.from_id AS artifact_id,
			operator.name AS operator_name,
		 	operator_result.execution_state as metadata,
			operator_result.workflow_dag_result_id 
		FROM workflow_dag_edge, operator, operator_result 
		WHERE 
			workflow_dag_edge.to_id = operator.id 
			AND operator.id = operator_result.operator_id 
			AND workflow_dag_edge.from_id IN (%s) 
			AND json_extract(operator.spec, '$.type') = '%s';`,
		stmt_preparers.GenerateArgsList(len(artifactIDs), 1),
		operator.CheckType,
	)
	args := stmt_preparers.CastIdsListToInterfaceList(artifactIDs)

	var statuses []views.OperatorResultStatus
	err := DB.Query(ctx, &statuses, query, args...)
	return statuses, err
}

func (*operatorResultReader) GetStatusByDAGResultAndArtifactBatch(
	ctx context.Context,
	dagResultIDs []uuid.UUID,
	artifactIDs []uuid.UUID,
	DB database.Database,
) ([]views.OperatorResultStatus, error) {
	// Get all unique artifact_id, execution_state, workflow_dag_result_id for all `workflow_dag_result_id`s
	// in `workflowDagResultIds` and `artifact_id`s in `artifactIds`.
	query := fmt.Sprintf(
		`SELECT DISTINCT 
			workflow_dag_edge.to_id AS artifact_id,
			operator_result.execution_state as metadata,
			operator_result.workflow_dag_result_id,
			NULL AS operator_name  
		FROM workflow_dag_edge, operator_result 
		WHERE 
			workflow_dag_edge.from_id = operator_result.operator_id 
			AND workflow_dag_edge.to_id IN (%s) 
			AND operator_result.workflow_dag_result_id IN (%s);`,
		stmt_preparers.GenerateArgsList(len(artifactIDs), 1),
		stmt_preparers.GenerateArgsList(len(dagResultIDs), len(artifactIDs)+1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(artifactIDs)
	args = append(args, stmt_preparers.CastIdsListToInterfaceList(dagResultIDs)...)

	var statuses []views.OperatorResultStatus
	err := DB.Query(ctx, &statuses, query, args...)
	return statuses, err
}

func (*operatorResultReader) GetOperatorWithArtifactResultNodesByOperatorNameAndWorkflow(
	ctx context.Context,
	operatorName string, 
	workflowID uuid.UUID,
	DB database.Database,
) ([]views.OperatorWithArtifactResultNode, error) {
	// For all workflow dags that belong to the workflow (identified by ID),
	// get the workflow dag edges of the workflow dag.
	// Get all operators with the operator name, get the operator ids and
	// find the operator results of each operator (by id).
	// Get all the artifact results by finding all workflow_dag_edges
	// from operator by id to artifact result by artifact id.
	query := `SELECT 
			operator_result.id,
			operator.id AS operator_id,
			operator_result.execution_state AS operator_result_exec_state,
			artifact_result.artifact_id, 
			artifact_result.id AS artifact_result_id, 
			artifact_result.metadata,
			artifact_result.content_path,
			artifact_result.execution_state AS artifact_result_exec_state,
		FROM operator, operator_result, artifact_result, workflow_dag, workflow_dag_edge
		WHERE 
			workflow_dag.workflow_id = $1
			AND workflow_dag_edge.workflow_dag_id = workflow_dag.id
			AND operator.name = $2
			AND operator_result.id = operator.id
			AND workflow_dag_edge.from_id = operator.id
			AND workflow_dag_edge.to_id = artifact_result.artifact_id
			AND artifact_result.workflow_dag_result_id = workflow_dag_result.id;`

	args := []interface{}{workflowID, operatorName}

	var operatorWithArtifactResultNodes []views.OperatorWithArtifactResultNode
	err := DB.Query(ctx, &operatorWithArtifactResultNodes, query, args...)
	return operatorWithArtifactResultNodes, err
}

func (*operatorResultWriter) Create(
	ctx context.Context,
	dagResultID uuid.UUID,
	operatorID uuid.UUID,
	execState *shared.ExecutionState,
	DB database.Database,
) (*models.OperatorResult, error) {
	cols := []string{
		models.OperatorResultID,
		models.OperatorResultDAGResultID,
		models.OperatorResultOperatorID,
		models.OperatorResultStatus,
		models.OperatorResultExecState,
	}
	query := DB.PrepareInsertWithReturnAllStmt(models.OperatorResultTable, cols, models.OperatorResultCols())

	ID, err := GenerateUniqueUUID(ctx, models.OperatorResultTable, DB)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		ID,
		dagResultID,
		operatorID,
		execState.Status,
		execState,
	}

	return getOperatorResult(ctx, DB, query, args...)
}

func (*operatorResultWriter) Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error {
	return deleteOperatorResults(ctx, DB, []uuid.UUID{ID})
}

func (*operatorResultWriter) DeleteBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) error {
	return deleteOperatorResults(ctx, DB, IDs)
}

func (*operatorResultWriter) Update(
	ctx context.Context,
	ID uuid.UUID,
	changes map[string]interface{},
	DB database.Database,
) (*models.OperatorResult, error) {
	var operatorResult models.OperatorResult
	err := repos.UpdateRecordToDest(
		ctx,
		&operatorResult,
		changes,
		models.OperatorResultTable,
		models.OperatorResultID,
		ID,
		models.OperatorResultCols(),
		DB,
	)
	return &operatorResult, err
}

func (*operatorResultWriter) UpdateBatchStatusByStatus(
	ctx context.Context,
	from shared.ExecutionStatus,
	to shared.ExecutionStatus,
	DB database.Database,
) ([]models.OperatorResult, error) {
	setExecStateFragment, args, err := generateUpdateExecStateSnippet(
		models.OperatorResultExecState,
		to,
		time.Now(),
		0, /* offset */
	)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		UPDATE %s SET
			%s,
			status = $%d
		WHERE
			json_extract(%s, '$.status') = $%d
		RETURNING %s;`,
		models.OperatorResultTable,
		setExecStateFragment,
		len(args)+1,
		models.OperatorResultExecState,
		len(args)+2,
		models.OperatorResultCols(),
	)

	args = append(args, to)
	args = append(args, from)
	var results []models.OperatorResult
	err = DB.Query(ctx, &results, query, args...)
	return results, err
}

func getOperatorResults(ctx context.Context, DB database.Database, query string, args ...interface{}) ([]models.OperatorResult, error) {
	var operatorResults []models.OperatorResult
	err := DB.Query(ctx, &operatorResults, query, args...)
	return operatorResults, err
}

func getOperatorResult(ctx context.Context, DB database.Database, query string, args ...interface{}) (*models.OperatorResult, error) {
	operatorResults, err := getOperatorResults(ctx, DB, query, args...)
	if err != nil {
		return nil, err
	}

	if len(operatorResults) == 0 {
		return nil, database.ErrNoRows()
	}

	if len(operatorResults) != 1 {
		return nil, errors.Newf("Expected 1 OperatorResult but got %v", len(operatorResults))
	}

	return &operatorResults[0], nil
}

func deleteOperatorResults(ctx context.Context, DB database.Database, IDs []uuid.UUID) error {
	if len(IDs) == 0 {
		return nil
	}

	query := fmt.Sprintf(
		`DELETE FROM operator_result WHERE id IN (%s)`,
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)

	return DB.Execute(ctx, query, args...)
}
