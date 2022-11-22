package shared

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/google/uuid"
)

type NotificationLevel string

const (
	SuccessLevel NotificationLevel = "success"
	WarningLevel NotificationLevel = "warning"
	ErrorLevel   NotificationLevel = "error"
	InfoLevel    NotificationLevel = "info"
	NeutralLevel NotificationLevel = "neutral"
)

type NotificationStatus string

const (
	UnreadStatus   NotificationStatus = "unread"
	ArchivedStatus NotificationStatus = "archived"
)

type NotificationObject string

const (
	WorkflowObject     NotificationObject = "workflow"
	DAGResultObject    NotificationObject = "workflow_dag_result"
	OrganizationObject NotificationObject = "organization"
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
