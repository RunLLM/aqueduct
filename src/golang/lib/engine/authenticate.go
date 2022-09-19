package engine

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/k8s"
	lambda_utils "github.com/aqueducthq/aqueduct/lib/lambda"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/dropbox/godropbox/errors"
)

const (
	TestFilePath = "src/"
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

func AuthenticateLambdaConfig(ctx context.Context, authConf auth.Config) error {
	lambdaConf, err := ParseLambdaConfig(authConf)
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

	for _, functionType := range functionsToShip {
		err := lambda_utils.CreateLambdaFunction(functionType, lambdaConf.RoleArn)
		if err != nil {
			return errors.Wrap(err, "Unable to Create Lambda Function")
		}
	}
	return nil
}
