package engine

import (
	"context"
	"time"

	"github.com/aqueducthq/aqueduct/lib/k8s"
	lambda_utils "github.com/aqueducthq/aqueduct/lib/lambda"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
	"golang.org/x/sync/errgroup"
)

// Authenticates kubernetes configuration by trying to connect a client.
func AuthenticateK8sConfig(ctx context.Context, authConf auth.Config) error {
	conf, err := ParseK8sConfig(authConf)
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
	lambdaConf, err := ParseLambdaConfig(authConf)
	if err != nil {
		return errors.Wrap(err, "Unable to parse configuration.")
	}

	errGroup := new(errgroup.Group)

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

	for _, functionType := range functionsToShip {
		lambdaFunctionType := functionType
		errGroup.Go(func() error {
			return lambda_utils.CreateLambdaFunction(lambdaFunctionType, lambdaConf.RoleArn)
		})
		time.Sleep(3 * time.Second)
	}

	if err := errGroup.Wait(); err != nil {
		return errors.Wrap(err, "Unable to Create Lambda Function.")
	}

	return nil
}
