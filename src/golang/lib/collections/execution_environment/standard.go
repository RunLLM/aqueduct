package execution_environment

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type standardReaderImpl struct{}

type standardWriterImpl struct{}

func (w *standardWriterImpl) CreateExecutionEnvironment(
	ctx context.Context,
	spec Spec, hash string,
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
		"DELETE FROM execution_environments WHERE id IN (%s);",
		stmt_preparers.GenerateArgsList(len(ids), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(ids)
	return db.Execute(ctx, deleteStmt, args...)
}
