package models

import (
	"strings"
	"time"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

const (
	NotificationTable = "notification"

	// Notification table column names
	NotificationID          = "id"
	NotificationReceiverID  = "receiver_id"
	NotificationContent     = "content"
	NotificationStatus      = "status"
	NotificationLevel       = "level"
	NotificationAssociation = "association"
	NotificationCreatedAt   = "created_at"
)

// A Notification maps to the notification table.
type Notification struct {
	ID          uuid.UUID                      `db:"id"`
	ReceiverID  uuid.UUID                      `db:"receiver_id"`
	Content     string                         `db:"content"`
	Status      shared.NotificationStatus      `db:"status"`
	Level       shared.NotificationLevel       `db:"level"`
	Association shared.NotificationAssociation `db:"association"`
	CreatedAt   time.Time                      `db:"created_at"`
}

// NotificationCols returns a comma-separated string of all Notification columns.
func NotificationCols() string {
	return strings.Join(allNotificationCols(), ",")
}

func allNotificationCols() []string {
	return []string{
		NotificationID,
		NotificationReceiverID,
		NotificationContent,
		NotificationStatus,
		NotificationLevel,
		NotificationAssociation,
		NotificationCreatedAt,
	}
}
