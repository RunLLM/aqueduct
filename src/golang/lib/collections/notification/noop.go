package notification

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type noopReaderImpl struct {
	throwError bool
}

type noopWriterImpl struct {
	throwError bool
}

func NewNoopReader(throwError bool) Reader {
	return &noopReaderImpl{throwError: throwError}
}

func NewNoopWriter(throwError bool) Writer {
	return &noopWriterImpl{throwError: throwError}
}

func (w *noopWriterImpl) CreateNotification(
	ctx context.Context,
	receiverId uuid.UUID,
	content string,
	level Level,
	association NotificationAssociation,
	db database.Database,
) (*Notification, error) {
	return nil, utils.NoopInterfaceErrorHandling(w.throwError)
}

func (r *noopReaderImpl) GetNotificationByReceiver(
	ctx context.Context,
	receiverId uuid.UUID,
	status Status,
	db database.Database,
) ([]Notification, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (w *noopWriterImpl) UpdateNotificationStatus(
	ctx context.Context,
	id uuid.UUID,
	status Status,
	db database.Database,
) (*Notification, error) {
	return nil, utils.NoopInterfaceErrorHandling(w.throwError)
}

func (r *noopReaderImpl) ValidateNotificationOwnership(
	ctx context.Context,
	notificationId uuid.UUID,
	userId uuid.UUID,
	db database.Database,
) (bool, error) {
	return true, nil
}
