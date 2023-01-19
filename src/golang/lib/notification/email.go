package notification

import (
	"crypto/tls"
	"net/smtp"

	"github.com/aqueducthq/aqueduct/lib/collections/integration"
)

func AuthenticateEmail(conf *integration.EmailConfig) error {
	auth := smtp.PlainAuth("", conf.User, conf.Password, conf.Host)
	client, err := smtp.Dial(conf.FullHost())
	if err != nil {
		return err
	}

	err = client.StartTLS(&tls.Config{ServerName: conf.Host})
	if err != nil {
		return err
	}

	defer client.Quit()
	return client.Auth(auth)
}
