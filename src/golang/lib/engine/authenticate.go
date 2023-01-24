package engine

import (
	"context"

	databricks_lib "github.com/aqueducthq/aqueduct/lib/databricks"
	"github.com/aqueducthq/aqueduct/lib/k8s"
	lambda_utils "github.com/aqueducthq/aqueduct/lib/lambda"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
	"golang.org/x/sync/errgroup"
)

const MaxConcurrentDownload = 3
const MaxConcurrentUpload = 5

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

	errGroup, _ := errgroup.WithContext(ctx)

	pullImageChannel := make(chan lambda_utils.LambdaFunctionType, MaxConcurrentDownload)

	pushImageChannel := make(chan lambda_utils.LambdaFunctionType, MaxConcurrentUpload)

	// Run authentication only once since the credentials will be recorded.
	err = lambda_utils.AuthenticateDockerToECR()
	if err != nil {
		return errors.Wrap(err, "Unable to authenticate Lambda Function.")
	}

	// Pull images on a currency of "MaxConcurrentDownload" to parallelize while avoiding pull timeout.
	errGroup.SetLimit(MaxConcurrentDownload)
	go AddFunctionTypeToChannel(functionsToShip, pullImageChannel)
	for lambdaFunction := range pullImageChannel {
		lambdaFunctionType := lambdaFunction
		errGroup.Go(func() error {
			return lambda_utils.PullImageFromECR(lambdaFunctionType)
		})
	}

	if err := errGroup.Wait(); err != nil {
		return errors.Wrap(err, "Unable to Create Lambda Function.")
	}

	// Push the images and create lambda functions all at once.
	errGroup.SetLimit(MaxConcurrentUpload)
	go AddFunctionTypeToChannel(functionsToShip, pushImageChannel)
	for lambdaFunction := range pushImageChannel {
		lambdaFunctionType := lambdaFunction
		errGroup.Go(func() error {
			return lambda_utils.CreateLambdaFunction(lambdaFunctionType, lambdaConf.RoleArn)
		})
	}

	if err := errGroup.Wait(); err != nil {
		return errors.Wrap(err, "Unable to Create Lambda Function.")
	}

	return nil
}

func AddFunctionTypeToChannel(functionsToShip [10]lambda_utils.LambdaFunctionType, channel chan lambda_utils.LambdaFunctionType) {
	for _, lambdaFunctionType := range functionsToShip {
		lambdaFunctionTypeToPass := lambdaFunctionType
		channel <- lambdaFunctionTypeToPass
	}
	close(channel)
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
