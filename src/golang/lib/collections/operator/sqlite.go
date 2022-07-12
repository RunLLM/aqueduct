package operator

import (
	"context"
	"fmt"

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

func (r *sqliteReaderImpl) GetOperatorsByIntegrationId(
	ctx context.Context,
	integrationId uuid.UUID,
	db database.Database,
) ([]Operator, error) {
	getOperatorsByIntegrationIdQuery := fmt.Sprintf(
		`SELET %s FROM %s
		WHERE json(spec)->>'$.load.integration_id' = $1
		AND json(spec)->>'$.extract.integration_id' = $2`,
		allColumns(),
		tableName,
	)

	var operators []Operator
	err := db.Query(
		ctx,
		&operators,
		getOperatorsByIntegrationIdQuery,
		integrationId,
		integrationId,
	)
	return operators, err
}

func (w *sqliteWriterImpl) CreateOperator(
	ctx context.Context,
	name string,
	description string,
	spec *Spec,
	db database.Database,
) (*Operator, error) {
	insertColumns := []string{IdColumn, NameColumn, DescriptionColumn, SpecColumn}
	insertOperatorStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	id, err := utils.GenerateUniqueUUID(ctx, tableName, db)
	if err != nil {
		return nil, err
	}

	args := []interface{}{id, name, description, spec}

	var operator Operator
	err = db.Query(ctx, &operator, insertOperatorStmt, args...)
	return &operator, err
}
