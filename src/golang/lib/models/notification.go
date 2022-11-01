package models

import (
	"time"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

// A Notification maps to the notification table.
type Notification struct {
	ID          uuid.UUID               `db:"id"`
	ReceiverId  uuid.UUID               `db:"receiver_id"`
	Content     string                  `db:"content"`
	Status      Status                  `db:"status"`
	Level       Level                   `db:"level"`
	Association NotificationAssociation `db:"association"`
	CreatedAt   time.Time               `db:"created_at"`
}