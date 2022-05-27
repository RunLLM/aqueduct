package operator_result

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

func (w *standardWriterImpl) CreateOperatorResult(
	ctx context.Context,
	workflowDagResultId uuid.UUID,
	operatorId uuid.UUID,
	db database.Database,
) (*OperatorResult, error) {
	insertColumns := []string{WorkflowDagResultIdColumn, OperatorIdColumn, StatusColumn}
	insertOperatorStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	args := []interface{}{workflowDagResultId, operatorId, shared.PendingExecutionStatus}

	var operatorResult OperatorResult
	err := db.Query(ctx, &operatorResult, insertOperatorStmt, args...)
	return &operatorResult, err
}

func (r *standardReaderImpl) GetOperatorResult(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) (*OperatorResult, error) {
	operatorResults, err := r.GetOperatorResults(ctx, []uuid.UUID{id}, db)
	if err != nil {
		return nil, err
	}

	if len(operatorResults) != 1 {
		return nil, errors.Newf("Expected 1 operator_result, but got %d operator_results.", len(operatorResults))
	}

	return &operatorResults[0], nil
}

func (r *standardReaderImpl) GetOperatorResults(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) ([]OperatorResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("Provided empty IDs list.")
	}

	getOperatorResultsQuery := fmt.Sprintf(
		"SELECT %s FROM operator_result WHERE id IN (%s);",
		allColumns(),
		stmt_preparers.GenerateArgsList(len(ids), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(ids)

	var operatorResults []OperatorResult
	err := db.Query(ctx, &operatorResults, getOperatorResultsQuery, args...)
	return operatorResults, err
}

func (r *standardReaderImpl) GetOperatorResultByWorkflowDagResultIdAndOperatorId(
	ctx context.Context,
	workflowDagResultId,
	operatorId uuid.UUID,
	db database.Database,
) (*OperatorResult, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM operator_result WHERE workflow_dag_result_id = $1 AND operator_id = $2;",
		allColumns(),
	)

	var operatorResult OperatorResult
	err := db.Query(ctx, &operatorResult, query, workflowDagResultId, operatorId)
	return &operatorResult, err
}

func (r *standardReaderImpl) GetOperatorResultsByWorkflowDagResultIds(
	ctx context.Context,
	workflowDagResultIds []uuid.UUID,
	db database.Database,
) ([]OperatorResult, error) {
	if len(workflowDagResultIds) == 0 {
		return nil, errors.New("Provided empty IDs list.")
	}

	query := fmt.Sprintf(
		"SELECT %s FROM operator_result WHERE workflow_dag_result_id IN (%s);",
		allColumns(),
		stmt_preparers.GenerateArgsList(len(workflowDagResultIds), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(workflowDagResultIds)

	var operatorResults []OperatorResult
	err := db.Query(ctx, &operatorResults, query, args...)
	return operatorResults, err
}

func (w *standardWriterImpl) UpdateOperatorResult(
	ctx context.Context,
	id uuid.UUID,
	changes map[string]interface{},
	db database.Database,
) (*OperatorResult, error) {
	var operatorResult OperatorResult
	err := utils.UpdateRecordToDest(ctx, &operatorResult, changes, tableName, IdColumn, id, allColumns(), db)
	return &operatorResult, err
}

func (w *standardWriterImpl) DeleteOperatorResult(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) error {
	return w.DeleteOperatorResults(ctx, []uuid.UUID{id}, db)
}

func (w *standardWriterImpl) DeleteOperatorResults(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) error {
	if len(ids) == 0 {
		return nil
	}

	deleteOperatorResultStmt := fmt.Sprintf(
		"DELETE FROM operator_result WHERE id IN (%s);",
		stmt_preparers.GenerateArgsList(len(ids), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(ids)
	return db.Execute(ctx, deleteOperatorResultStmt, args...)
}
