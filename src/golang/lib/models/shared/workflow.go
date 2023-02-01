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
	return utils.ScanJSONB(value, s)
}

type NullNotificationSettings struct {
	NotificationSettings
	IsNull bool
}

func (n *NullNotificationSettings) Value() (driver.Value, error) {
	if n.IsNull {
		return nil, nil
	}

	return (&n.NotificationSettings).Value()
}

func (n *NullNotificationSettings) Scan(value interface{}) error {
	if value == nil {
		n.IsNull = true
		return nil
	}

	s := &NotificationSettings{}
	if err := s.Scan(value); err != nil {
		return err
	}

	n.NotificationSettings, n.IsNull = *s, false
	return nil
}
