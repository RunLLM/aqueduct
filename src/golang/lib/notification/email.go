package notification

import (
	"crypto/tls"
	"net/smtp"

	"github.com/aqueducthq/aqueduct/lib/collections/integration"
)

func AuthenticateEmail(conf *integration.EmailConfig) error {
	// Reference: https://gist.github.com/jim3ma/b5c9edeac77ac92157f8f8affa290f45
	auth := smtp.PlainAuth("", conf.User, conf.Password, conf.Host)
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
