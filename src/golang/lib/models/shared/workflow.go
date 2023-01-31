package shared

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/google/uuid"
)

// `NotificationSettings` maps IntegrationID to NotificationLevel
// This has to be a struct since sql driver does not support map type.
type NotificationSettings struct {
	Settings map[uuid.UUID]NotificationLevel `json:"settings"`
}

func (s *NotificationSettings) Value() (driver.Value, error) {
	return utils.ValueJSONB(*s)
}

func (s *NotificationSettings) Scan(value interface{}) error {
	if value == nil {
		s.Settings = nil
		return nil
	}

	return utils.ScanJSONB(value, s)
}
