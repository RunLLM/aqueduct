package integration

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/database"
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

func (r *postgresReaderImpl) GetIntegrationsByConfigField(
	ctx context.Context,
	fieldName string,
	fieldValue string,
	db database.Database,
) ([]Integration, error) {
	getIntegrationsQuery := fmt.Sprintf(
		"SELECT %s FROM integration WHERE json_extract_text(config, $1) = $2;",
		allColumns(),
	)
	var integrations []Integration

	err := db.Query(ctx, &integrations, getIntegrationsQuery, fieldName, fieldValue)
	return integrations, err
}
