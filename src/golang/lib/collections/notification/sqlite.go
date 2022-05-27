package notification

import (
	"context"
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

func (w *sqliteWriterImpl) CreateNotification(
	ctx context.Context,
	receiverId uuid.UUID,
	content string,
	level Level,
	association NotificationAssociation,
	db database.Database,
) (*Notification, error) {
	insertColumns := []string{
		IdColumn,
		ReceiverIdColumn,
		ContentColumn,
		StatusColumn,
		LevelColumn,
		AssociationColumn,
		CreatedAtColumn,
	}
	insertStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	id, err := utils.GenerateUniqueUUID(ctx, tableName, db)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		id, receiverId, content, UnreadStatus, level, &association, time.Now(),
	}

	var notification Notification
	err = db.Query(ctx, &notification, insertStmt, args...)
	return &notification, err
}
