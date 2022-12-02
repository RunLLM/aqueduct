package shared

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/google/uuid"
)

type NotificationLevel string

const (
	SuccessNotificationLevel NotificationLevel = "success"
	WarningNotificationLevel NotificationLevel = "warning"
	ErrorNotificationLevel   NotificationLevel = "error"
	InfoNotificationLevel    NotificationLevel = "info"
	NeutralNotificationLevel NotificationLevel = "neutral"
)

type NotificationStatus string

const (
	UnreadNotificationStatus   NotificationStatus = "unread"
	ArchivedNotificationStatus NotificationStatus = "archived"
)

type NotificationObject string

const (
	WorkflowNotificationObject  NotificationObject = "workflow"
	DAGResultNotificationObject NotificationObject = "workflow_dag_result"
	OrgNotificationObject       NotificationObject = "organization"
)

type NotificationAssociation struct {
	Object NotificationObject `json:"object"`
	ID     uuid.UUID          `json:"id"`
}

func (association *NotificationAssociation) Value() (driver.Value, error) {
	return utils.ValueJSONB(*association)
}

func (association *NotificationAssociation) Scan(value interface{}) error {
	return utils.ScanJSONB(value, association)
}
