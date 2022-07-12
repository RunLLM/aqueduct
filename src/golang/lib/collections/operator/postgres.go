package operator

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type postgresReaderImpl struct {
	standardReaderImpl
}

type postgresWriterImpl struct {
	standardWriterImpl
}

func newPostgresReader() Reader {
	return &postgresReaderImpl{standardReaderImpl{}}
}

func newPostgresWriter() Writer {
	return &postgresWriterImpl{standardWriterImpl{}}
}

func (r *postgresReaderImpl) GetOperatorsByIntegrationId(
	ctx context.Context,
	integrationId uuid.UUID,
	db database.Database,
) ([]Operator, error) {
	getOperatorsByIntegrationIdQuery := fmt.Sprintf(
		`SELET %s FROM %s
		WHERE spec->load->>'integration_id' = $1
		AND spec->extract->>'integration_id' = $2`,
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
