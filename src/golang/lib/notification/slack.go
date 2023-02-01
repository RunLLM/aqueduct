package notification

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	"github.com/slack-go/slack"
)

const maxChannelLimit = 2000

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

// reference: https://stackoverflow.com/questions/50106263/slack-api-to-find-existing-channel
// We have to use list channel API together with a linear search.
func findChannels(client *slack.Client, names []string) ([]slack.Channel, error) {
	channels, _ /* cursor */, err := client.GetConversations(&slack.GetConversationsParameters{
		ExcludeArchived: true,
		Limit:           maxChannelLimit,
	})
	if err != nil {
		return nil, err
	}

	namesSet := make(map[string]bool, len(names))
	namesTaken := make(map[string]bool, len(names))
	for _, name := range names {
		namesSet[name] = true
	}

	// Slack channel names should be unique, based on
	// https://twitter.com/slackhq/status/965358920146608128?lang=en .
	// However, we make this list dynamic size on purpose to prevent potential
	// compatibility issue in future. We will still send notifications to all
	// channels matching the given name.
	results := []slack.Channel{}
	for _, channel := range channels {
		_, ok := namesSet[channel.Name]
		if ok {
			results = append(results, channel)
			namesTaken[channel.Name] = true
		}
	}

	if len(namesTaken) != len(namesSet) {
		for name := range namesSet {
			_, ok := namesTaken[name]
			if !ok {
				return nil, errors.Newf("Channel %s does not exist.", name)
			}
		}
	}

	return results, nil
}

func (s *SlackNotification) Send(ctx context.Context, msg string) error {
	client := slack.New(s.conf.Token)
	channels, err := findChannels(client, s.conf.Channels)
	if err != nil {
		return err
	}

	for _, channel := range channels {
		// reference: https://medium.com/@gausha/a-simple-slackbot-with-golang-c5a932d719c7
		_, _, _, err = client.SendMessage(channel.ID, slack.MsgOptionBlocks(
			slack.NewSectionBlock(
				slack.NewTextBlockObject(
					"plain_text",
					msg,
					false, /* emoji */
					false, /* verbatim */
				),
				nil,
				nil,
			),
		))

		if err != nil {
			return err
		}
	}

	return nil
}

func AuthenticateSlack(conf *shared.SlackConfig) error {
	client := slack.New(conf.Token)
	_, err := client.AuthTest()
	if err != nil {
		return err
	}

	_, err = findChannels(client, conf.Channels)
	return err
}
