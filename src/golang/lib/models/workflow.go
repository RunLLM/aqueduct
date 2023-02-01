package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

const (
	WorkflowTable = "workflow"

	// Workflow column names
	WorkflowID                   = "id"
	WorkflowUserID               = "user_id"
	WorkflowName                 = "name"
	WorkflowDescription          = "description"
	WorkflowSchedule             = "schedule"
	WorkflowCreatedAt            = "created_at"
	WorkflowRetentionPolicy      = "retention_policy"
	WorkflowNotificationSettings = "notification_settings"
)

// A Workflow maps to the workflow table.
type Workflow struct {
	ID                   uuid.UUID                       `db:"id" json:"id"`
	UserID               uuid.UUID                       `db:"user_id" json:"user_id"`
	Name                 string                          `db:"name" json:"name"`
	Description          string                          `db:"description" json:"description"`
	Schedule             workflow.Schedule               `db:"schedule" json:"schedule"`
	CreatedAt            time.Time                       `db:"created_at" json:"created_at"`
	RetentionPolicy      workflow.RetentionPolicy        `db:"retention_policy" json:"retention_policy"`
	NotificationSettings shared.NullNotificationSettings `db:"notification_settings" json:"notification_settings"`
}

// WorkflowCols returns a comma-separated string of all Workflow columns.
func WorkflowCols() string {
	return strings.Join(allWorkflowCols(), ",")
}

// WorkflowColsWithPrefix returns a comma-separated string of all
// Workflow columns prefixed by the table name.
func WorkflowColsWithPrefix() string {
	cols := allWorkflowCols()
	for i, col := range cols {
		cols[i] = fmt.Sprintf("%s.%s", WorkflowTable, col)
	}

	return strings.Join(cols, ",")
}

func allWorkflowCols() []string {
	return []string{
		WorkflowID,
		WorkflowUserID,
		WorkflowName,
		WorkflowDescription,
		WorkflowSchedule,
		WorkflowCreatedAt,
		WorkflowRetentionPolicy,
		WorkflowNotificationSettings,
	}
}
