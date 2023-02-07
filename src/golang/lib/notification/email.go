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
	fullMsg += fmt.Sprintf("Content-Type: text/html; charset=\"UTF-8\";\n\n")
	fullMsg += body
	return fullMsg
}

func (e *EmailNotification) SendForDag(
	ctx context.Context,
	wfDag dag.WorkflowDag,
	level shared.NotificationLevel,
	contextMsg string,
) error {
	subject := summary(wfDag, level)
	contextBlock := ""
	if contextMsg != "" {
		contextBlock = fmt.Sprintf(`<div>
			<b>Context</b>:
		</div>
		<div>
			<font face="monospace">%s</font>
		</div>`, contextMsg)
	}
	body := fmt.Sprintf(`<div dir="ltr">
		<b>Result ID</b>: <font face="monospace">%s</font>
		%s
	</div>`, wfDag.ResultID(), contextBlock)
	fullMsg := fullMessage(subject, e.conf.User, e.conf.Targets, body)

	return e.send(ctx, fullMsg)
}

func (e *EmailNotification) send(ctx context.Context, msg string) error {
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
