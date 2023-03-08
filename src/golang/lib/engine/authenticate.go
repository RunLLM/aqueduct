package engine

import (
	"context"

	databricks_lib "github.com/aqueducthq/aqueduct/lib/databricks"
	"github.com/aqueducthq/aqueduct/lib/k8s"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/spark"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
)

// Authenticates kubernetes configuration by trying to connect a client.
func AuthenticateK8sConfig(ctx context.Context, authConf auth.Config) error {
	conf, err := lib_utils.ParseK8sConfig(authConf)
	if err != nil {
		return errors.Wrap(err, "Unable to parse configuration.")
	}

	if conf.Dynamic {
		if conf.CloudIntegrationId == "" {
			return errors.New("Dynamic K8s integration must have a cloud integration ID attached.")
		} else {
			return nil
		}
	}

	_, err = k8s.CreateK8sClient(conf.KubeconfigPath, bool(conf.UseSameCluster))
	if err != nil {
		return errors.Wrap(err, "Unable to create kubernetes client.")
	}
	return nil
}

func AuthenticateDatabricksConfig(ctx context.Context, authConf auth.Config) error {
	databricksConfig, err := lib_utils.ParseDatabricksConfig(authConf)
	if err != nil {
		return errors.Wrap(err, "Unable to parse configuration.")
	}

	databricksClient, err := databricks_lib.NewWorkspaceClient(
		databricksConfig.WorkspaceURL,
		databricksConfig.AccessToken,
	)
	if err != nil {
		return errors.Wrap(err, "Unable to create Databricks Workspace Client.")
	}
	_, err = databricks_lib.ListJobs(
		ctx,
		databricksClient,
	)
	if err != nil {
		return errors.Wrap(err, "Unable to list Databricks Jobs.")
	}

	err = databricks_lib.AddEntrypointFilesToStorage(ctx)
	if err != nil {
		return errors.Wrap(err, "Unable to upload entrypoint files to storage.")
	}

	return nil
}

func AuthenticateSparkConfig(ctx context.Context, authConf auth.Config) error {
	sparkConfig, err := lib_utils.ParseSparkConfig(authConf)
	if err != nil {
		return errors.Wrap(err, "Unable to parse configuration.")
	}

	livyClient := spark.NewLivyClient(sparkConfig.LivyServerURL)
	_, err = livyClient.GetSessions()
	if err != nil {
		return errors.Wrap(err, "Unable to list active Sessions on Livy Server.")
	}

	return nil
}

func AuthenticateAWSConfig(authConf auth.Config) error {
	conf, err := lib_utils.ParseAWSConfig(authConf)
	if err != nil {
		return errors.Wrap(err, "Unable to parse configuration.")
	}

	if conf.AccessKeyId == "" || conf.SecretAccessKey == "" {
		return errors.New("AWS access key ID and secret access key must be provided.")
	}

	return nil
}
