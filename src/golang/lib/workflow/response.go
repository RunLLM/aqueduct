package workflow

import (
	"time"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

// This file should map exactly to
// `src/ui/common/src/handlers/responses/workflow.ts`
type Response struct {
	ID                   uuid.UUID                   `json:"id"`
	UserID               uuid.UUID                   `json:"user_id"`
	Name                 string                      `json:"name"`
	Description          string                      `json:"description"`
	Schedule             shared.Schedule             `json:"schedule"`
	CreatedAt            time.Time                   `json:"created_at"`
	RetentionPolicy      shared.RetentionPolicy      `json:"retention_policy"`
	NotificationSettings shared.NotificationSettings `json:"notification_settings"`
}

func NewResponseFromDBObject(dbWorkflow *models.Workflow) *Response {
	return &Response{
		ID:                   dbWorkflow.ID,
		UserID:               dbWorkflow.UserID,
		Name:                 dbWorkflow.Name,
		Description:          dbWorkflow.Description,
		Schedule:             dbWorkflow.Schedule,
		CreatedAt:            dbWorkflow.CreatedAt,
		RetentionPolicy:      dbWorkflow.RetentionPolicy,
		NotificationSettings: dbWorkflow.NotificationSettings,
	}
}
