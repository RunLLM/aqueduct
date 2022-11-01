package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/google/uuid"
)

// Notification defines all of the database operations that can be performed for a Notification.
type Notification interface {
	notificationReader
	notificationWriter
}

type notificationReader interface {
	// Get returns the Workflow with id.
	GetByReceiver(ctx context.Context, receiverId uuid.UUID, status Status, db database.Database) ([]models.Notification, error)

	// ValidateUser returns whether the Notification belongs to userID.
	ValidateUser(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID, db database.Database) (bool, error)
}

type notificationWriter interface {
	// Create inserts a new Notification with the specified fields.
	Create(
		ctx context.Context,
		receiverId uuid.UUID,
		content string,
		level Level,
		association NotificationAssociation,
		db database.Database,
	) (*models.Notification, error)

	// Update applies changes to the status of the Notification with id. It returns the updated Notification.
	Update(ctx context.Context, id uuid.UUID, status Status, db database.Database) (*models.Notification, error)
}
