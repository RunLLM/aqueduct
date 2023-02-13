package notification

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

var ErrIntegrationTypeIsNotNotification = errors.New("Integration type is not a notification.")

type Notification interface {
	// `ID()` is the unique identifier, typically mapped to the integration ID.
	ID() uuid.UUID

	// `Level()` is the global default severity level threshold beyond which a notification should send.
	// For example, 'warning' threshold allows 'error' and 'warning' level notifications,
	// but blocking 'success' notifications.
	//
	// This behavior is controlled by caller calling `ShouldSend()` function.
	// This field is a 'global default' as we allow overriding this behavior in,
	// for example, workflow specific settings.
	Level() shared.NotificationLevel

	// `Send()` sends a notification.
	// The caller should decide, based on `Level()` and any other context, if `Send()`
	// should be called.
	Send(ctx context.Context, msg string) error
}

func GetNotificationsFromUser(
	ctx context.Context,
	userID uuid.UUID,
	integrationRepo repos.Integration,
	vaultObject vault.Vault,
	DB database.Database,
) ([]Notification, error) {
	emailIntegrations, err := integrationRepo.GetByServiceAndUser(ctx, shared.Email, userID, DB)
	if err != nil {
		return nil, err
	}

	slackIntegrations, err := integrationRepo.GetByServiceAndUser(ctx, shared.Slack, userID, DB)
	if err != nil {
		return nil, err
	}

	allIntegrations := make([]models.Integration, 0, len(emailIntegrations)+len(slackIntegrations))
	allIntegrations = append(allIntegrations, emailIntegrations...)
	allIntegrations = append(allIntegrations, slackIntegrations...)
	notifications := make([]Notification, 0, len(allIntegrations))
	for _, integrationObj := range allIntegrations {
		integrationCopied := integrationObj
		notification, err := NewNotificationFromIntegration(ctx, &integrationCopied, vaultObject)
		if err != nil {
			return nil, err
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}

func NewNotificationFromIntegration(
	ctx context.Context,
	integrationObject *models.Integration,
	vaultObject vault.Vault,
) (Notification, error) {
	if integrationObject.Service == shared.Email {
		conf, err := auth.ReadConfigFromSecret(ctx, integrationObject.ID, vaultObject)
		if err != nil {
			return nil, err
		}

		emailConf, err := lib_utils.ParseEmailConfig(conf)
		if err != nil {
			return nil, err
		}

		return newEmailNotification(integrationObject, emailConf), nil
	}

	if integrationObject.Service == shared.Slack {
		conf, err := auth.ReadConfigFromSecret(ctx, integrationObject.ID, vaultObject)
		if err != nil {
			return nil, err
		}

		slackConf, err := lib_utils.ParseSlackConfig(conf)
		if err != nil {
			return nil, err
		}

		return newSlackNotification(integrationObject, slackConf), nil
	}

	return nil, ErrIntegrationTypeIsNotNotification
}

// `ShouldSend` determines if a notification at 'level' passes configuration
// specified by `thresholdLevel`.
// 'info' and 'neutral' will get through regardless of threshold.
// And 'info' or 'neutral' threshold lets everything through.
// Other states will follow the severity ordering.
func ShouldSend(
	thresholdLevel shared.NotificationLevel,
	level shared.NotificationLevel,
) bool {
	if thresholdLevel == shared.InfoNotificationLevel || thresholdLevel == shared.NeutralNotificationLevel {
		return true
	}

	levelSeverityMap := map[shared.NotificationLevel]int{
		shared.SuccessNotificationLevel: 0,
		shared.WarningNotificationLevel: 1,
		shared.ErrorNotificationLevel:   2,
		shared.InfoNotificationLevel:    3,
		shared.NeutralNotificationLevel: 3,
	}

	return levelSeverityMap[level] >= levelSeverityMap[thresholdLevel]
}
