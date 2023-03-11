package handler

import (
	"context"
	"net/http"

	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/google/uuid"
)

// Route: /notifications
// Method: GET
// Params: None
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//
// Response:
//
//	Body:
//		serialized `listNotificationsResponse`, a list of notifications for the user
type ListNotificationsHandler struct {
	GetHandler

	Database database.Database

	DAGResultRepo    repos.DAGResult
	NotificationRepo repos.Notification
}

type listNotificationsResponse []notificationResponse

type notificationResponse struct {
	ID                        uuid.UUID                       `json:"id"`
	Content                   string                          `json:"content"`
	Status                    shared.NotificationStatus       `json:"status"`
	Level                     shared.NotificationLevel        `json:"level"`
	Association               shared.NotificationAssociation  `json:"association"`
	CreatedAt                 int64                           `json:"createdAt"`
	DAGResultWorkflowMetadata views.DAGResultWorkflowMetadata `json:"workflowMetadata"`
}

func (*ListNotificationsHandler) Name() string {
	return "ListNotifications"
}

func (*ListNotificationsHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	return aqContext, http.StatusOK, nil
}

func (h *ListNotificationsHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*aq_context.AqContext)

	emptyResponse := listNotificationsResponse{}

	// For now, we hard-code to retrieve all notifications with 'unread' status.
	// This API can be extended in future to handle reading notifications with other types, or status.
	notifications, err := h.NotificationRepo.GetByReceiverAndStatus(
		ctx,
		args.ID,
		shared.UnreadNotificationStatus,
		h.Database,
	)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to list notifications.")
	}

	dagResultIDs := make([]uuid.UUID, 0, len(notifications))
	for _, notification := range notifications {
		dagResultIDs = append(dagResultIDs, notification.Association.ID)
	}

	dagResultToWorkflowMetadata := make(map[uuid.UUID]views.DAGResultWorkflowMetadata, len(dagResultIDs))
	if len(dagResultIDs) > 0 {
		dagResultToWorkflowMetadata, err = h.DAGResultRepo.GetWorkflowMetadataBatch(ctx, dagResultIDs, h.Database)
		if err != nil {
			return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve workflow info related to workflow dag result.")
		}
	}

	responses := make([]notificationResponse, 0, len(notifications))
	for _, notificationObject := range notifications {
		responses = append(responses, notificationResponse{
			ID:                        notificationObject.ID,
			Content:                   notificationObject.Content,
			Status:                    notificationObject.Status,
			Level:                     notificationObject.Level,
			Association:               notificationObject.Association,
			CreatedAt:                 notificationObject.CreatedAt.Unix(),
			DAGResultWorkflowMetadata: dagResultToWorkflowMetadata[notificationObject.Association.ID],
		})
	}

	return responses, http.StatusOK, nil
}
