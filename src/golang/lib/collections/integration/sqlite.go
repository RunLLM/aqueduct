package integration

import (
	"context"
	"fmt"
	"time"

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

func (w *sqliteWriterImpl) CreateIntegration(
	ctx context.Context,
	organizationId string,
	service Service,
	name string,
	config *utils.Config,
	validated bool,
	db database.Database,
) (*Integration, error) {
	insertColumns := []string{
		IdColumn, OrganizationIdColumn, ServiceColumn, NameColumn, ConfigColumn, CreatedAtColumn, ValidatedColumn,
	}
	insertIntegrationStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	id, err := utils.GenerateUniqueUUID(ctx, tableName, db)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		id, organizationId, service, name, config, time.Now(), validated,
	}

	var integration Integration
	err = db.Query(ctx, &integration, insertIntegrationStmt, args...)
	return &integration, err
}

func (w *sqliteWriterImpl) CreateIntegrationForUser(
	ctx context.Context,
	organizationId string,
	userId uuid.UUID,
	service Service,
	name string,
	config *utils.Config,
	validated bool,
	db database.Database,
) (*Integration, error) {
	insertColumns := []string{
		IdColumn, OrganizationIdColumn, UserIdColumn, ServiceColumn,
		NameColumn, ConfigColumn, CreatedAtColumn, ValidatedColumn,
	}
	insertIntegrationStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	id, err := utils.GenerateUniqueUUID(ctx, tableName, db)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		id, organizationId, userId, service, name, config, time.Now(), validated,
	}

	var integration Integration
	err = db.Query(ctx, &integration, insertIntegrationStmt, args...)
	return &integration, err
}

func (r *sqliteReaderImpl) GetIntegrationsByConfigField(
	ctx context.Context,
	fieldName string,
	fieldValue string,
	db database.Database,
) ([]Integration, error) {
	getIntegrationsQuery := fmt.Sprintf(
		"SELECT %s FROM integration WHERE json_extract(config, $1) = $2;",
		allColumns(),
	)
	var integrations []Integration

	// The full 'where' condition becomes
	// `json_extract(config, '$.field_name') = 'field_value'`
	// which matches https://www.sqlite.org/json1.html .
	// We parametrize the extracted field_name and field_value
	// to prevent injection.
	err := db.Query(ctx, &integrations, getIntegrationsQuery, "$."+fieldName, fieldValue)
	return integrations, err
}
