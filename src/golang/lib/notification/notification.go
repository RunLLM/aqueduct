package notification

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
	"github.com/gofrs/uuid"
)

var ErrIntegrationTypeIsNotNotification = errors.New("Integration type is not a notification.")

type Notification interface {
	// `Send()` sends a notification with `level`, and the content is `msg`.
	//
	// The caller always call `Send()` when a notification is generated.
	// There could be a level preference associated with the notification integration.
	// For example, slack and email has `level` field in config,
	// and only notifications beyond this level will be sent.
	// In such cases, the implementation of `Send()` should reflect the level preference.
	Send(msg string, level shared.NotificationLevel) error
}

func GetNotificationsFromUser(
	ctx context.Context,
	userID uuid.UUID,
	integrationRepo repos.Integration,
	vaultObject vault.Vault,
	DB database.Database,
) ([]Notification, error) {
	emailIntegrations, err := integrationRepo.GetByServiceAndUser(ctx, integration.Email, userID, DB)
	if err != nil {
		return nil, err
	}

	slackIntegrations, err := integrationRepo.GetByServiceAndUser(ctx, integration.Slack, userID, DB)
	if err != nil {
		return nil, err
	}

	allIntegrations := make([]Notification, 0, len(emailIntegrations)+len(slackIntegrations))
	allIntegrations = append(allIntegrations, emailIntegrations...)
	allIntegrations = append(allIntegrations, slackIntegrations...)
	notifications := make([]Notification, 0, len(allIntegrations))
	for _, integrationObj := range allIntegrations {
		notification, err := NewNotificationFromIntegration(ctx, integrationObj, vaultObject)
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
	if integrationObject.Service == integration.Email {
		conf, err := auth.ReadConfigFromSecret(ctx, integrationObject.ID, vaultObject)
		if err != nil {
			return nil, err
		}

		emailConf, err := lib_utils.ParseEmailConfig(conf)
		if err != nil {
			return nil, err
		}

		return newEmailNotification(emailConf), nil
	}

	if integrationObject.Service == integration.Slack {
		conf, err := auth.ReadConfigFromSecret(ctx, integrationObject.ID, vaultObject)
		if err != nil {
			return nil, err
		}

		slackConf, err := lib_utils.ParseSlackConfig(conf)
		if err != nil {
			return nil, err
		}

		return newSlackNotification(slackConf), nil
	}

	return nil, ErrIntegrationTypeIsNotNotification
}
