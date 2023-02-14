package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type notificationRepo struct {
	notificationReader
	notificationWriter
}

type notificationReader struct{}

type notificationWriter struct{}

func NewNotificationRepo() repos.Notification {
	return &notificationRepo{
		notificationReader: notificationReader{},
		notificationWriter: notificationWriter{},
	}
}

func (*notificationReader) GetByReceiverAndStatus(ctx context.Context, receiverID uuid.UUID, status shared.NotificationStatus, DB database.Database) ([]models.Notification, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM notification WHERE receiver_id = $1 AND status = $2;`,
		models.NotificationCols(),
	)
	args := []interface{}{receiverID, status}

	return getNotifications(ctx, DB, query, args...)
}

func (*notificationReader) ValidateUser(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID, DB database.Database) (bool, error) {
	query := `SELECT COUNT(*) AS count FROM notification WHERE id = $1 AND receiver_id = $2;`
	var count countResult

	err := DB.Query(ctx, &count, query, notificationID, userID)
	if err != nil {
		return false, err
	}

	return count.Count == 1, nil
}

func (*notificationWriter) Create(
	ctx context.Context,
	receiverID uuid.UUID,
	content string,
	level shared.NotificationLevel,
	association *shared.NotificationAssociation,
	DB database.Database,
) (*models.Notification, error) {
	cols := []string{
		models.NotificationID,
		models.NotificationReceiverID,
		models.NotificationContent,
		models.NotificationStatus,
		models.NotificationLevel,
		models.NotificationAssociation,
		models.NotificationCreatedAt,
	}
	query := DB.PrepareInsertWithReturnAllStmt(models.NotificationTable, cols, models.NotificationCols())

	ID, err := utils.GenerateUniqueUUID(ctx, models.UserTable, DB)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		ID,
		receiverID,
		content,
		shared.UnreadNotificationStatus,
		level,
		association,
		time.Now(),
	}
	return getNotification(ctx, DB, query, args...)
}

func (*notificationWriter) Update(ctx context.Context, ID uuid.UUID, status shared.NotificationStatus, DB database.Database) (*models.Notification, error) {
	changedColumns := map[string]interface{}{
		models.NotificationStatus: status,
	}
	return updateNotification(ctx, ID, changedColumns, DB)
}

func updateNotification(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, DB database.Database) (*models.Notification, error) {
	var notification models.Notification
	err := utils.UpdateRecordToDest(ctx, &notification, changes, models.NotificationTable, models.NotificationID, ID, models.NotificationCols(), DB)
	return &notification, err
}

func getNotifications(ctx context.Context, DB database.Database, query string, args ...interface{}) ([]models.Notification, error) {
	var notifications []models.Notification
	err := DB.Query(ctx, &notifications, query, args...)
	return notifications, err
}

func getNotification(ctx context.Context, DB database.Database, query string, args ...interface{}) (*models.Notification, error) {
	notifications, err := getNotifications(ctx, DB, query, args...)
	if err != nil {
		return nil, err
	}

	if len(notifications) == 0 {
		return nil, database.ErrNoRows
	}

	if len(notifications) != 1 {
		return nil, errors.Newf("Expected 1 notification but got %v", len(notifications))
	}

	return &notifications[0], nil
}
