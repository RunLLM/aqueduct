package notification

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/workflow/dag"
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

func (s *SlackNotification) constructOperatorMessages(wfDag dag.WorkflowDag) string {
	warningOps := wfDag.OperatorsWithWarning()
	errorOps := wfDag.OperatorsWithError()

	if len(warningOps)+len(errorOps) == 0 {
		return ""
	}

	// there are at least some checks failed:
	msg := "\n"
	for _, op := range warningOps {
		msg += fmt.Sprintf(
			"%s `%s` failed (warning).\n",
			constructDisplayedOperatorType(op.Type()),
			op.Name(),
		)
	}

	for _, op := range errorOps {
		msg += fmt.Sprintf(
			"%s `%s` failed (error).\n",
			constructDisplayedOperatorType(op.Type()),
			op.Name(),
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

	link := wfDag.ResultLink()
	linkWarning := ""
	linkWarningStr := constructLinkWarning(link)
	if len(linkWarningStr) > 0 {
		linkWarning = fmt.Sprintf("(%s)", linkWarningStr)
	}

	linkContent := fmt.Sprintf("See the Aqueduct UI for more details: %s %s", link, linkWarning)
	nameContent := fmt.Sprintf("*Workflow:* `%s`", wfDag.Name())
	IDContent := fmt.Sprintf("*ID:* `%s`", wfDag.ID())
	resultIDContent := fmt.Sprintf("*Result ID:* `%s`", wfDag.ResultID())
	msg := fmt.Sprintf(
		"%s\n%s\n%s%s%s\n%s",
		nameContent,
		IDContent,
		resultIDContent,
		s.constructOperatorMessages(wfDag),
		contextMarkdownBlock,
		linkContent,
	)
	for _, channel := range channels {
		// reference: https://medium.com/@gausha/a-simple-slackbot-with-golang-c5a932d719c7
		_, _, _, err = client.SendMessageContext(ctx, channel.ID, slack.MsgOptionBlocks(
			slack.NewHeaderBlock(
				slack.NewTextBlockObject(
					"plain_text",
					summarize(wfDag, level),
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
