package schema_version

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
)

type standardReaderImpl struct{}

type standardWriterImpl struct{}

func (w *standardWriterImpl) CreateSchemaVersion(
	ctx context.Context,
	version int64,
	name string,
	db database.Database,
) (*SchemaVersion, error) {
	insertColumns := []string{VersionColumn, DirtyColumn, NameColumn}
	insertSchemaVersionStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	args := []interface{}{version, true, name}

	var schemaVersion SchemaVersion
	err := db.Query(ctx, &schemaVersion, insertSchemaVersionStmt, args...)
	return &schemaVersion, err
}

func (r *standardReaderImpl) GetSchemaVersion(
	ctx context.Context,
	version int64,
	db database.Database,
) (*SchemaVersion, error) {
	getSchemaVersionQuery := fmt.Sprintf(
		"SELECT %s FROM schema_version WHERE version = %d;",
		allColumns(),
		version,
	)

	var schemaVersion SchemaVersion
	err := db.Query(ctx, &schemaVersion, getSchemaVersionQuery)
	return &schemaVersion, err
}

func (r *standardReaderImpl) GetCurrentSchemaVersion(
	ctx context.Context,
	db database.Database,
) (*SchemaVersion, error) {
	getCurrentSchemaVersionQuery := fmt.Sprintf(
		"SELECT %s FROM schema_version ORDER BY version DESC LIMIT 1;",
		allColumns(),
	)

	var schemaVersion SchemaVersion
	err := db.Query(ctx, &schemaVersion, getCurrentSchemaVersionQuery)
	return &schemaVersion, err
}

func (w *standardWriterImpl) UpdateSchemaVersion(
	ctx context.Context,
	version int64,
	changes map[string]interface{},
	db database.Database,
) (*SchemaVersion, error) {
	var schemaVersion SchemaVersion
	err := utils.UpdateRecordToDest(ctx, &schemaVersion, changes, tableName, VersionColumn, version, allColumns(), db)
	return &schemaVersion, err
}

func (w *standardWriterImpl) DeleteSchemaVersion(
	ctx context.Context,
	version int64,
	db database.Database,
) error {
	deleteSchemaVersionStmt := fmt.Sprintf("DELETE FROM schema_version WHERE version = %d;", version)
	return db.Execute(ctx, deleteSchemaVersionStmt)
}
