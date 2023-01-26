package notification

import (
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/slack-go/slack"
)

func AuthenticateSlack(conf *shared.SlackConfig) error {
	client := slack.New(conf.Token)
	_, err := client.AuthTest()
	return err
}
