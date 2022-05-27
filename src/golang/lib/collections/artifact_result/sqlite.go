package artifact_result

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

func (w *sqliteWriterImpl) CreateArtifactResult(
	ctx context.Context,
	workflowDagResultId uuid.UUID,
	artifactId uuid.UUID,
	contentPath string,
	db database.Database,
) (*ArtifactResult, error) {
	insertColumns := []string{IdColumn, WorkflowDagResultIdColumn, ArtifactIdColumn, ContentPathColumn, StatusColumn}
	insertArtifactStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	id, err := utils.GenerateUniqueUUID(ctx, tableName, db)
	if err != nil {
		return nil, err
	}

	args := []interface{}{id, workflowDagResultId, artifactId, contentPath, shared.PendingExecutionStatus}

	var artifactResult ArtifactResult
	err = db.Query(ctx, &artifactResult, insertArtifactStmt, args...)
	return &artifactResult, err
}
