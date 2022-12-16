package sqlite

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
)

type schemaVersionRepo struct {
	schemaVersionReader
	schemaVersionWriter
}

type schemaVersionReader struct{}

type schemaVersionWriter struct{}

func NewSchemaVersionRepo() repos.SchemaVersion {
	return &schemaVersionRepo{
		schemaVersionReader: schemaVersionReader{},
		schemaVersionWriter: schemaVersionWriter{},
	}
}

func (*schemaVersionReader) Get(ctx context.Context, version int64, DB database.Database) (*models.SchemaVersion, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM schema_version WHERE version = $1`,
		models.SchemaVersionCols(),
	)
	args := []interface{}{version}

	return getSchemaVersion(ctx, DB, query, args...)
}

func (*schemaVersionReader) GetCurrent(ctx context.Context, DB database.Database) (*models.SchemaVersion, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM schema_version ORDER BY version DESC LIMIT 1;`,
		models.SchemaVersionCols(),
	)

	return getSchemaVersion(ctx, DB, query)
}

func (*schemaVersionWriter) Create(
	ctx context.Context,
	version int64,
	name string,
	DB database.Database,
) (*models.SchemaVersion, error) {
	cols := []string{
		models.SchemaVersionVersion,
		models.SchemaVersionName,
		models.SchemaVersionDirty,
	}
	query := DB.PrepareInsertWithReturnAllStmt(models.SchemaVersionTable, cols, models.SchemaVersionCols())

	args := []interface{}{
		version,
		name,
		true,
	}
	return getSchemaVersion(ctx, DB, query, args...)
}

func (*schemaVersionWriter) Delete(ctx context.Context, version int64, DB database.Database) error {
	query := `DELETE FROM schema_version WHERE version = $1;`
	args := []interface{}{version}

	return DB.Execute(ctx, query, args...)
}

func (*schemaVersionWriter) Update(ctx context.Context, version int64, changes map[string]interface{}, DB database.Database) (*models.SchemaVersion, error) {
	var schemaVersion models.SchemaVersion
	err := utils.UpdateRecordToDest(
		ctx,
		&schemaVersion,
		changes,
		models.SchemaVersionTable,
		models.SchemaVersionVersion,
		version,
		models.SchemaVersionCols(),
		DB,
	)
	return &schemaVersion, err
}

func getSchemaVersions(ctx context.Context, DB database.Database, query string, args ...interface{}) ([]models.SchemaVersion, error) {
	var schemaVersions []models.SchemaVersion
	err := DB.Query(ctx, &schemaVersions, query, args...)
	return schemaVersions, err
}

func getSchemaVersion(ctx context.Context, DB database.Database, query string, args ...interface{}) (*models.SchemaVersion, error) {
	schemaVersions, err := getSchemaVersions(ctx, DB, query, args...)
	if err != nil {
		return nil, err
	}

	if len(schemaVersions) == 0 {
		return nil, database.ErrNoRows
	}

	if len(schemaVersions) != 1 {
		return nil, errors.Newf("Expected 1 schemaVersion but got %v", len(schemaVersions))
	}

	return &schemaVersions[0], nil
}
