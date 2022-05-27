package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type standardReaderImpl struct{}

type standardWriterImpl struct{}

func (w *standardWriterImpl) CreateNotification(
	ctx context.Context,
	receiverId uuid.UUID,
	content string,
	level Level,
	association NotificationAssociation,
	db database.Database,
) (*Notification, error) {
	insertColumns := []string{
		ReceiverIdColumn,
		ContentColumn,
		StatusColumn,
		LevelColumn,
		AssociationColumn,
		CreatedAtColumn,
	}
	insertStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	args := []interface{}{
		receiverId, content, UnreadStatus, level, association, time.Now(),
	}

	var notification Notification
	err := db.Query(ctx, &notification, insertStmt, args...)
	return &notification, err
}

func (r *standardReaderImpl) GetNotificationByReceiver(
	ctx context.Context,
	receiverId uuid.UUID,
	status Status,
	db database.Database,
) ([]Notification, error) {
	getNotificationsQuery := fmt.Sprintf(
		"SELECT %s FROM notification WHERE receiver_id = $1 AND status = $2;",
		allColumns(),
	)
	var notifications []Notification

	err := db.Query(ctx, &notifications, getNotificationsQuery, receiverId, status)
	return notifications, err
}

func updateNotification(
	ctx context.Context,
	id uuid.UUID,
	changedColumns map[string]interface{},
	db database.Database,
) (*Notification, error) {
	var notification Notification
	err := utils.UpdateRecordToDest(ctx, &notification, changedColumns, tableName, IdColumn, id, allColumns(), db)
	return &notification, err
}

func (w *standardWriterImpl) UpdateNotificationStatus(
	ctx context.Context,
	id uuid.UUID,
	status Status,
	db database.Database,
) (*Notification, error) {
	changedColumns := map[string]interface{}{
		StatusColumn: status,
	}

	return updateNotification(ctx, id, changedColumns, db)
}

func (r *standardReaderImpl) ValidateNotificationOwnership(
	ctx context.Context,
	notificationId uuid.UUID,
	userId uuid.UUID,
	db database.Database,
) (bool, error) {
	query := `SELECT COUNT(*) AS count FROM notification WHERE id = $1 AND receiver_id = $2;`
	var count utils.CountResult

	err := db.Query(ctx, &count, query, notificationId, userId)
	if err != nil {
		return false, err
	}

	return count.Count == 1, nil
}
