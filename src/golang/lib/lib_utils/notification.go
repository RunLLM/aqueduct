package lib_utils

import (
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
)

const (
	notificationEnabledKey = "enabled"
	notificationLevelKey   = "level"
)

// Returning a nil level means that `disabled` == true.
func ExtractNotificationLevel(integrationObject *models.Resource) (*shared.NotificationLevel, error) {
	enabledStr, ok := integrationObject.Config[notificationEnabledKey]
	if !ok {
		return nil, errors.Newf("Notification %s is missing 'enabled' key.", integrationObject.Name)
	}
	if enabledStr == "false" {
		return nil, nil
	}

	levelStr, ok := integrationObject.Config[notificationLevelKey]
	if !ok {
		return nil, errors.Newf("Notification %s is enabled but missing 'level' key.", integrationObject.Name)
	}
	level, err := shared.StrToNotificationLevel(levelStr)
	if err != nil {
		return nil, err
	}
	return &level, nil
}
