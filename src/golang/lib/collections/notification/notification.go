package notification

import (
	"context"
	"time"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type Notification struct {
	Id          uuid.UUID               `db:"id"`
	ReceiverId  uuid.UUID               `db:"receiver_id"`
	Content     string                  `db:"content"`
	Status      Status                  `db:"status"`
	Level       Level                   `db:"level"`
	Association NotificationAssociation `db:"association"`
	CreatedAt   time.Time               `db:"created_at"`
}

type Reader interface {
	GetNotificationByReceiver(
		ctx context.Context,
		receiverId uuid.UUID,
		status Status,
		db database.Database,
	) ([]Notification, error)
	ValidateNotificationOwnership(
		ctx context.Context,
		notificationId uuid.UUID,
		userId uuid.UUID,
		db database.Database,
	) (bool, error)
}

type Writer interface {
	CreateNotification(
		ctx context.Context,
		receiverId uuid.UUID,
		content string,
		level Level,
		association NotificationAssociation,
		db database.Database,
	) (*Notification, error)
	UpdateNotificationStatus(
		ctx context.Context,
		id uuid.UUID,
		status Status,
		db database.Database,
	) (*Notification, error)
}

func NewReader(dbConf *database.DatabaseConfig) (Reader, error) {
	if dbConf.Type == database.PostgresType {
		return newPostgresReader(), nil
	}

	if dbConf.Type == database.SqliteType {
		return newSqliteReader(), nil
	}

	return nil, database.ErrUnsupportedDbType
}

func NewWriter(dbConf *database.DatabaseConfig) (Writer, error) {
	if dbConf.Type == database.PostgresType {
		return newPostgresWriter(), nil
	}

	if dbConf.Type == database.SqliteType {
		return newSqliteWriter(), nil
	}

	return nil, database.ErrUnsupportedDbType
}
