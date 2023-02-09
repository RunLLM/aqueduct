package engine

import (
	"context"
	"time"

	databricks_lib "github.com/aqueducthq/aqueduct/lib/databricks"
	"github.com/aqueducthq/aqueduct/lib/k8s"
	lambda_utils "github.com/aqueducthq/aqueduct/lib/lambda"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	MaxConcurrentDownload = 3
	MaxConcurrentUpload   = 5
)

// Authenticates kubernetes configuration by trying to connect a client.
func AuthenticateK8sConfig(ctx context.Context, authConf auth.Config) error {
	conf, err := lib_utils.ParseK8sConfig(authConf)
	if err != nil {
		return errors.Wrap(err, "Unable to parse configuration.")
	}
	_, err = k8s.CreateK8sClient(conf.KubeconfigPath, bool(conf.UseSameCluster))
	if err != nil {
		return errors.Wrap(err, "Unable to create kubernetes client.")
	}
	return nil
}

func AuthenticateLambdaConfig(ctx context.Context, authConf auth.Config) error {
	lambdaConf, err := lib_utils.ParseLambdaConfig(authConf)
	if err != nil {
		return errors.Wrap(err, "Unable to parse configuration.")
	}
	functionsToShip := [10]lambda_utils.LambdaFunctionType{
		lambda_utils.FunctionExecutor37Type,
		lambda_utils.FunctionExecutor38Type,
		lambda_utils.FunctionExecutor39Type,
		lambda_utils.ParamExecutorType,
		lambda_utils.SystemMetricType,
		lambda_utils.AthenaConnectorType,
		lambda_utils.BigQueryConnectorType,
		lambda_utils.PostgresConnectorType,
		lambda_utils.S3ConnectorType,
		lambda_utils.SnowflakeConnectorType,
	}

	// Run authentication only once since the credentials will be recorded.
	err = lambda_utils.AuthenticateDockerToECR()
	if err != nil {
		return errors.Wrap(err, "Unable to authenticate Lambda Function.")
	}
	// Pull images on a concurrency of "MaxConcurrentDownload".

	errGroup, _ := errgroup.WithContext(ctx)

	pullImageChannel := make(chan lambda_utils.LambdaFunctionType, len(functionsToShip))
	defer close(pullImageChannel)
	pushImageChannel := make(chan lambda_utils.LambdaFunctionType, len(functionsToShip))
	defer close(pushImageChannel)
	lambda_utils.AddFunctionTypeToChannel(functionsToShip[:], pullImageChannel)
	for i := 0; i < MaxConcurrentDownload; i++ {
		errGroup.Go(func() error {
			for {
				select {
				case functionType := <-pullImageChannel:
					lambdaFunctionType := functionType
					log.Info("Pulling", lambdaFunctionType)
					err := lambda_utils.PullImageFromECR(lambdaFunctionType)
					if err != nil {
						return err
					}
					pushImageChannel <- functionType
				default:
					log.Info("stop here ", len(pullImageChannel))
					return nil
				}
			}
		})
	}

	// Create lambda functions on a concurrency of "MaxConcurrentUpload".

	signalChannel := make(chan lambda_utils.LambdaFunctionType, len(functionsToShip))
	lambda_utils.AddFunctionTypeToChannel(functionsToShip[:], signalChannel)

	for i := 0; i < MaxConcurrentUpload; i++ {
		errGroup.Go(func() error {
			for {
				select {
				case functionType := <-pushImageChannel:
					lambdaFunctionType := functionType
					log.Info("Pushing", lambdaFunctionType)
					err := lambda_utils.CreateLambdaFunction(lambdaFunctionType, lambdaConf.RoleArn)
					if err != nil {
						return err
					}
					<-signalChannel
				default:
					time.Sleep(1 * time.Second)
					if len(signalChannel) == 0 {
						return nil
					}
				}
			}
		})
	}

	if err := errGroup.Wait(); err != nil {
		return errors.Wrap(err, "Unable to Create Lambda Function.")
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
