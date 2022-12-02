package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/dropbox/godropbox/errors"
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

	Database           database.Database
	NotificationReader notification.Reader
	// TODO: Replace this with repos.Workflow after notification refactor
	WorkflowReader workflow.Reader
}

type listNotificationsResponse []notificationResponse

type notificationResponse struct {
	Id               uuid.UUID                             `json:"id"`
	Content          string                                `json:"content"`
	Status           notification.Status                   `json:"status"`
	Level            notification.Level                    `json:"level"`
	Association      notification.NotificationAssociation  `json:"association"`
	CreatedAt        int64                                 `json:"createdAt"`
	WorkflowMetadata workflow.NotificationWorkflowMetadata `json:"workflowMetadata"`
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
	notifications, err := h.NotificationReader.GetNotificationByReceiver(
		ctx,
		args.ID,
		notification.UnreadStatus,
		h.Database,
	)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to list notifications.")
	}

	workflowDagResultIds := make([]uuid.UUID, 0, len(notifications))
	for _, notification := range notifications {
		workflowDagResultIds = append(workflowDagResultIds, notification.Association.Id)
	}

	workflowsMetadataMap := make(map[uuid.UUID]workflow.NotificationWorkflowMetadata)
	if len(workflowDagResultIds) > 0 {
		workflowsMetadataMap, err = h.WorkflowReader.GetNotificationWorkflowMetadata(ctx, workflowDagResultIds, h.Database)
		if err != nil {
			return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve workflow info related to workflow dag result.")
		}
	}

	responses := make([]notificationResponse, 0, len(notifications))
	for _, notificationObject := range notifications {
		responses = append(responses, notificationResponse{
			Id:               notificationObject.Id,
			Content:          notificationObject.Content,
			Status:           notificationObject.Status,
			Level:            notificationObject.Level,
			Association:      notificationObject.Association,
			CreatedAt:        notificationObject.CreatedAt.Unix(),
			WorkflowMetadata: workflowsMetadataMap[notificationObject.Association.Id],
		})
	}

	return responses, http.StatusOK, nil
}
