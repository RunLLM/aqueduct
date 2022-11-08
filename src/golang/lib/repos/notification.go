package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

// Notification defines all of the database operations that can be performed for a Notification.
type Notification interface {
	notificationReader
	notificationWriter
}

type notificationReader interface {
	// GetByReceiver returns the Workflow with ID.
	GetByReceiver(ctx context.Context, receiverID uuid.UUID, status shared.Status, db database.Database) ([]models.Notification, error)

	// ValidateUser returns whether the Notification belongs to userID.
	ValidateUser(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID, db database.Database) (bool, error)
}

type notificationWriter interface {
	// Create inserts a new Notification with the specified fields.
	Create(
		ctx context.Context,
		receiverID uuid.UUID,
		content string,
		level shared.Level,
		association shared.NotificationAssociation,
		db database.Database,
	) (*models.Notification, error)

	// Update applies changes to the status of the Notification with ID. It returns the updated Notification.
	Update(ctx context.Context, ID uuid.UUID, status shared.Status, db database.Database) (*models.Notification, error)
}
