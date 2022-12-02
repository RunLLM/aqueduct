package execution_environment

import (
	"context"

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

func (w *sqliteWriterImpl) CreateExecutionEnvironment(
	ctx context.Context,
	spec *Spec,
	hash uuid.UUID,
	db database.Database,
) (*DBExecutionEnvironment, error) {
	insertColumns := []string{
		IdColumn,
		SpecColumn,
		HashColumn,
	}
	insertStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	id, err := utils.GenerateUniqueUUID(ctx, tableName, db)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		id, spec, hash,
	}

	var result DBExecutionEnvironment
	err = db.Query(ctx, &result, insertStmt, args...)
	return &result, err
}
