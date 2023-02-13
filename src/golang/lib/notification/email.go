package notification

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/workflow/dag"
	"github.com/google/uuid"
)

type EmailNotification struct {
	integration *models.Integration
	conf        *shared.EmailConfig
}

func newEmailNotification(integration *models.Integration, conf *shared.EmailConfig) *EmailNotification {
	return &EmailNotification{integration: integration, conf: conf}
}

func (e *EmailNotification) ID() uuid.UUID {
	return e.integration.ID
}

func (e *EmailNotification) Level() shared.NotificationLevel {
	return e.conf.Level
}

func fullMessage(subject string, from string, targets []string, body string) string {
	fullMsg := fmt.Sprintf("From: %s\n", from)
	fullMsg += fmt.Sprintf("To: %s\n", strings.Join(targets, ","))
	fullMsg += fmt.Sprintf("Subject: %s\n", subject)
	fullMsg += "Content-Type: text/html; charset=\"UTF-8\";\n\n"
	fullMsg += body
	return fullMsg
}

func (e *EmailNotification) constructOperatorMessages(wfDag dag.WorkflowDag) string {
	warningOps := wfDag.OperatorsWithWarning()
	errorOps := wfDag.OperatorsWithError()

	if len(warningOps)+len(errorOps) == 0 {
		return ""
	}

	// there are at least some operators failed:
	msg := ""
	for _, op := range warningOps {
		msg += fmt.Sprintf(
			`<div>%s <font face="monospace">%s</font> failed (warning).</div>`,
			constructDisplayedOperatorType(op.Type()),
			op.Name(),
		)
	}

	for _, op := range errorOps {
		msg += fmt.Sprintf(
			`<div>%s <font face="monospace">%s</font> failed (error).</div>`,
			constructDisplayedOperatorType(op.Type()),
			op.Name(),
		)
	}

	return msg
}

func (e *EmailNotification) SendForDag(
	ctx context.Context,
	wfDag dag.WorkflowDag,
	level shared.NotificationLevel,
	systemErrContext string,
) error {
	subject := summarize(wfDag, level)
	systemErrBlock := ""
	if systemErrContext != "" {
		systemErrBlock = fmt.Sprintf(`<div>
			<b>Error:</b>
		</div>
		<div>
			<font face="monospace">%s</font>
		</div>`, systemErrContext)
	}

	body := fmt.Sprintf(`<div dir="ltr">
		<div>Go to Aqueduct UI for more details: <a href="%s">%s</a></div>
		<div><b>Workflow</b>: <font face="monospace">%s</font></div>
		<div><b>ID</b>: <font face="monospace">%s</font></div>
		<div><b>Result ID</b>: <font face="monospace">%s</font></div>
		%s
		%s
		</div>`,
		wfDag.ResultLink(),
		wfDag.ResultLink(),
		wfDag.Name(),
		wfDag.ID(),
		wfDag.ResultID(),
		e.constructOperatorMessages(wfDag),
		systemErrBlock,
	)
	fullMsg := fullMessage(subject, e.conf.User, e.conf.Targets, body)

	return e.send(fullMsg)
}

func (e *EmailNotification) send(msg string) error {
	auth := smtp.PlainAuth(
		"", // identity
		e.conf.User,
		e.conf.Password,
		e.conf.Host,
	)

	return smtp.SendMail(
		e.conf.FullHost(),
		auth,
		e.conf.User,
		e.conf.Targets,
		[]byte(msg),
	)
}

func AuthenticateEmail(conf *shared.EmailConfig) error {
	// Reference: https://gist.github.com/jim3ma/b5c9edeac77ac92157f8f8affa290f45
	auth := smtp.PlainAuth(
		"", // identity
		conf.User,
		conf.Password,
		conf.Host,
	)
	client, err := smtp.Dial(conf.FullHost())
	if err != nil {
		return err
	}

	err = client.StartTLS(&tls.Config{
		ServerName: conf.Host,
		// Reference: https://github.com/go-redis/redis/issues/1553
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return err
	}

	defer client.Close()
	return client.Auth(auth)
}
