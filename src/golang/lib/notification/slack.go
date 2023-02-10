package notification

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/workflow/dag"
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

func (s *SlackNotification) Enabled() bool {
	return s.conf.Enabled
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

	// Slack channel names should be unique. We will still send notifications to all
	// channels matching the given name.
	results := make([]slack.Channel, 0, len(names))
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

func (s *SlackNotification) checkMessages(wfDag dag.WorkflowDag) string {
	warningChecks := wfDag.ChecksWithWarning()
	errorChecks := wfDag.ChecksWithError()

	if len(warningChecks)+len(errorChecks) == 0 {
		return ""
	}

	// there are at least some checks failed:
	msg := "\n"
	for _, check := range warningChecks {
		msg += fmt.Sprintf(
			"Check `%s` failed (warning).\n",
			check.Name(),
		)
	}

	for _, check := range errorChecks {
		msg += fmt.Sprintf(
			"Check `%s` failed (error).\n",
			check.Name(),
		)
	}

	return msg
}

func (s *SlackNotification) SendForDag(
	ctx context.Context,
	wfDag dag.WorkflowDag,
	level shared.NotificationLevel,
	systemErrContext string,
) error {
	client := slack.New(s.conf.Token)
	channels, err := findChannels(client, s.conf.Channels)
	if err != nil {
		return err
	}

	contextMarkdownBlock := ""
	if systemErrContext != "" {
		contextMarkdownBlock = fmt.Sprintf("\n*Error:*\n%s", systemErrContext)
	}

	linkContent := fmt.Sprintf("Check Aqueduct UI for more details: %s", wfDag.ResultLink())
	nameContent := fmt.Sprintf("*Workflow:* `%s`", wfDag.Name())
	IDContent := fmt.Sprintf("*ID:* `%s`", wfDag.ID())
	resultIDContent := fmt.Sprintf("*Result ID:* `%s`", wfDag.ResultID())
	msg := fmt.Sprintf(
		"%s\n%s\n%s\n%s%s%s",
		linkContent,
		nameContent,
		IDContent,
		resultIDContent,
		s.checkMessages(wfDag),
		contextMarkdownBlock,
	)
	for _, channel := range channels {
		// reference: https://medium.com/@gausha/a-simple-slackbot-with-golang-c5a932d719c7
		_, _, _, err = client.SendMessageContext(ctx, channel.ID, slack.MsgOptionBlocks(
			slack.NewHeaderBlock(
				slack.NewTextBlockObject(
					"plain_text",
					summary(wfDag, level),
					false,
					false,
				),
			),
			slack.NewSectionBlock(
				slack.NewTextBlockObject(
					"mrkdwn",
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
