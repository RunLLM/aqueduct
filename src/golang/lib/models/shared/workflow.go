package shared

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/google/uuid"
)

// `NotificationSettings` maps IntegrationID to NotificationLevel
type NotificationSettings map[uuid.UUID]NotificationLevel

func (s *NotificationSettings) Value() (driver.Value, error) {
	return utils.ValueJSONB(*s)
}

func (s *NotificationSettings) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	return utils.ScanJSONB(value, s)
}
