package notification

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
	"github.com/slack-go/slack"
)

type SlackNotification struct {
	integration *models.Integration
	conf        *shared.SlackConfig
}

func newSlackNotification(integration *models.Integration, conf *shared.SlackConfig) *SlackNotification {
	return &SlackNotification{integration: integration, conf: conf}
}

func (s *SlackNotification) ID() uuid.UUID {
	return s.integration.ID
}

func (s *SlackNotification) Level() shared.NotificationLevel {
	return s.conf.Level
}

func (s *SlackNotification) Send(ctx context.Context, msg string) error {
	// TODO: Implement
	return nil
}

func AuthenticateSlack(conf *shared.SlackConfig) error {
	client := slack.New(conf.Token)
	_, err := client.AuthTest()
	return err
}
