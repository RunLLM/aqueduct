package sqlite

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
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

func (*notificationReader) GetByReceiver(ctx context.Context, receiverID uuid.UUID, status shared.NotificationStatus, DB database.Database) ([]models.Notification, error) {
	
}

func (*notificationReader) ValidateUser(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID, DB database.Database) (bool, error) {

}

func (*notificationWriter) Create(
	ctx context.Context,
	receiverID uuid.UUID,
	content string,
	level shared.NotificationLevel,
	association shared.NotificationAssociation,
	DB database.Database,
) (*models.Notification, error) {
	
}

func (*notificationWriter) Update(ctx context.Context, ID uuid.UUID, status shared.NotificationStatus, DB database.Database) (*models.Notification, error) {
	
}