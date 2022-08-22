package operator_result

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type sqliteReaderImpl struct {
	standardReaderImpl
}

type sqliteWriterImpl struct {
	standardWriterImpl
}

func newSqliteReader() Reader {
	return &sqliteReaderImpl{standardReaderImpl{}}
}

func newSqliteWriter() Writer {
	return &sqliteWriterImpl{standardWriterImpl{}}
}

func (w *sqliteWriterImpl) CreateOperatorResult(
	ctx context.Context,
	workflowDagResultId uuid.UUID,
	operatorId uuid.UUID,
	db database.Database,
) (*OperatorResult, error) {
	insertColumns := []string{IdColumn, WorkflowDagResultIdColumn, OperatorIdColumn, StatusColumn}
	insertOperatorStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	id, err := utils.GenerateUniqueUUID(ctx, tableName, db)
	if err != nil {
		return nil, err
	}

	args := []interface{}{id, workflowDagResultId, operatorId, shared.PendingExecutionStatus}

	var operatorResult OperatorResult
	err = db.Query(ctx, &operatorResult, insertOperatorStmt, args...)
	return &operatorResult, err
}

func (w *sqliteWriterImpl) InsertOperatorResult(
	ctx context.Context,
	workflowDagResultId uuid.UUID,
	operatorId uuid.UUID,
	execState *shared.ExecutionState,
	db database.Database,
) (*OperatorResult, error) {
	insertColumns := []string{IdColumn, WorkflowDagResultIdColumn, OperatorIdColumn, StatusColumn, ExecStateColumn}
	insertOperatorStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	id, err := utils.GenerateUniqueUUID(ctx, tableName, db)
	if err != nil {
		return nil, err
	}

	args := []interface{}{id, workflowDagResultId, operatorId, execState.Status, execState}

	var operatorResult OperatorResult
	err = db.Query(ctx, &operatorResult, insertOperatorStmt, args...)
	return &operatorResult, err
}
