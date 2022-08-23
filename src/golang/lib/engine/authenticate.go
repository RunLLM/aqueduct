package engine

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/k8s"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
)

// Authenticates kubernetes configuration by trying to connect a client.
func AuthenticateK8sConfig(ctx context.Context, authConf auth.Config) error {
	conf, err := ParseK8sConfig(authConf)
	if err != nil {
		return errors.Wrap(err, "Unable to parse configuration.")
	}
	_, err = k8s.CreateClientOutsideCluster(conf.KubeconfigPath)
	if err != nil {
		return errors.Wrap(err, "Unable to create kubernetes client.")
	}
	return nil
}
