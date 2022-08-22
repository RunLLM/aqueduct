package artifact

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
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

func (w *sqliteWriterImpl) CreateArtifact(
	ctx context.Context,
	name string,
	description string,
	artifactType Type,
	db database.Database,
) (*DBArtifact, error) {
	insertColumns := []string{IdColumn, NameColumn, DescriptionColumn, TypeColumn}
	insertArtifactStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	id, err := utils.GenerateUniqueUUID(ctx, tableName, db)
	if err != nil {
		return nil, err
	}

	args := []interface{}{id, name, description, artifactType}

	var artifact DBArtifact
	err = db.Query(ctx, &artifact, insertArtifactStmt, args...)
	return &artifact, err
}
