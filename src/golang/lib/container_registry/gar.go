package container_registry

import (
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
)

func AuthenticateGARConfig(authConf auth.Config) error {
	conf, err := lib_utils.ParseGARConfig(authConf)
	if err != nil {
		return errors.Wrap(err, "Unable to parse configuration.")
	}

	if conf.ServiceAccountKey == "" {
		return errors.New("Service account key is empty.")
	}

	return nil
}
