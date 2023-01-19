package notification

import (
	"net/smtp"

	"github.com/aqueducthq/aqueduct/lib/collections/integration"
)

func AuthenticateEmail(conf *integration.EmailConfig) error {
	auth := smtp.PlainAuth("", conf.User, conf.Password, conf.FullHost())
	client, err := smtp.Dial(conf.FullHost())
	if err != nil {
		return err
	}

	defer client.Close()
	return client.Auth(auth)
}
