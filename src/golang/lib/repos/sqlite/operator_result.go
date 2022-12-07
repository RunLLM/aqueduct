package sqlite

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/aqueducthq/aqueduct/lib/models"
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

func (*operatorResultReader) GetByDAGResultBatch(
	ctx context.Context,
	dagResultIDs []uuid.UUID,
	DB database.Database,
) ([]models.OperatorResult, error) {
	query := fmt.Sprintf(
		`SELECT %s 
		FROM operator_result 
		WHERE workflow_dag_result IN (%s);`,
		models.OperatorResultCols(),
		stmt_preparers.GenerateArgsList(len(dagResultIDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(dagResultIDs)

	return getOperatorResults(ctx, DB, query, args...)
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

	ID, err := utils.GenerateUniqueUUID(ctx, models.OperatorResultTable, DB)
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
	err := utils.UpdateRecordToDest(
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
		return nil, database.ErrNoRows
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
