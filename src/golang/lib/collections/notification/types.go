package notification

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/google/uuid"
)

type Status string

const (
	UnreadStatus   Status = "unread"
	ArchivedStatus Status = "archived"
)

type Level string

const (
	SuccessLevel Level = "success"
	WarningLevel Level = "warning"
	ErrorLevel   Level = "error"
	InfoLevel    Level = "info"
	NeutralLevel Level = "neutral"
)

type Object string

const (
	WorkflowObject          Object = "workflow"
	WorkflowDagResultObject Object = "workflow_dag_result"
	OrganizationObject      Object = "organization"
)

type NotificationAssociation struct {
	Object Object    `json:"object"`
	Id     uuid.UUID `json:"id"`
}

func (association *NotificationAssociation) Value() (driver.Value, error) {
	return utils.ValueJsonB(*association)
}

func (association *NotificationAssociation) Scan(value interface{}) error {
	return utils.ScanJsonB(value, association)
}
