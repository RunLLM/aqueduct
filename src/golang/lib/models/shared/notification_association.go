package shared

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/google/uuid"
)

type NotificationAssociation struct {
	Object Object    `json:"object"`
	Id     uuid.UUID `json:"id"`
}

func (association *NotificationAssociation) Value() (driver.Value, error) {
	return utils.ValueJSONB(*association)
}

func (association *NotificationAssociation) Scan(value interface{}) error {
	return utils.ScanJSONB(value, association)
}
