package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Route: /notifications/{notificationId}/archive
// Method: POST
// Params: notificationId
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//
// Response: none
type ArchiveNotificationHandler struct {
	PostHandler

	Database database.Database

	NotificationRepo repos.Notification
}

type archiveNotificationArgs struct {
	notificationID uuid.UUID
	userID         uuid.UUID
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

	notificationIDStr := chi.URLParam(r, routes.NotificationIdUrlParam)
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed notification ID.")
	}

	ok, err := h.NotificationRepo.ValidateUser(r.Context(), notificationID, aqContext.ID, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during notification ownership validation.")
	}

	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "This notification does not belong to the user.")
	}

	return &archiveNotificationArgs{
		notificationID: notificationID,
		userID:         aqContext.ID,
	}, http.StatusOK, nil
}

func (h *ArchiveNotificationHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*archiveNotificationArgs)
	emptyResp := archiveNotificationResponse{}

	_, err := h.NotificationRepo.Update(
		ctx,
		args.notificationID,
		shared.ArchivedNotificationStatus,
		h.Database,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to archive notification.")
	}

	return emptyResp, http.StatusOK, nil
}
