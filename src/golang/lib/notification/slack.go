package notification

import (
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/slack-go/slack"
)

type SlackNotification struct {
	conf *shared.SlackConfig
}

func newSlackNotification(conf *shared.SlackConfig) *SlackNotification {
	return &SlackNotification{conf: conf}
}

func (e *SlackNotification) Send(msg string, level shared.NotificationLevel) error {
	// TODO: Implement
	return nil
}

func AuthenticateSlack(conf *shared.SlackConfig) error {
	client := slack.New(conf.Token)
	_, err := client.AuthTest()
	return err
}
