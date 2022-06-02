package server

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

// Route: /notifications/{notificationId}/archive
// Method: POST
// Params: notificationId
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response: none
type ArchiveNotificationHandler struct {
	PostHandler

	NotificationReader notification.Reader
	NotificationWriter notification.Writer
	Database           database.Database
}

type archiveNotificationArgs struct {
	notificationId uuid.UUID
	userId         uuid.UUID
}

type archiveNotificationResponse struct{}

func (*ArchiveNotificationHandler) Name() string {
	return "ArchiveNotification"
}

func (h *ArchiveNotificationHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statuscode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statuscode, err
	}

	notificationIdStr := chi.URLParam(r, routes.NotificationIdUrlParam)
	notificationId, err := uuid.Parse(notificationIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed notification ID.")
	}

	ok, err := h.NotificationReader.ValidateNotificationOwnership(r.Context(), notificationId, aqContext.Id, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during notification ownership validation.")
	}

	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "This notification does not belong to the user.")
	}

	return &archiveNotificationArgs{
		notificationId: notificationId,
		userId:         aqContext.Id,
	}, http.StatusOK, nil
}

func (h *ArchiveNotificationHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*archiveNotificationArgs)
	emptyResp := archiveNotificationResponse{}

	_, err := h.NotificationWriter.UpdateNotificationStatus(
		ctx,
		args.notificationId,
		notification.ArchivedStatus,
		h.Database,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to archive notification.")
	}

	return emptyResp, http.StatusOK, nil
}
