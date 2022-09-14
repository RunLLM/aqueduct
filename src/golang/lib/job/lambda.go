package job

import (
	"context"
	"encoding/json"

	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	lambda_utils "github.com/aqueducthq/aqueduct/lib/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/dropbox/godropbox/errors"
)

const (
	defaultLambdaFunctionExtractPath = "/tmp/app/function/"
)

type lambdaJobManager struct {
	lambdaService *lambda.Lambda
	conf          *LambdaJobManagerConfig
}

func NewLambdaJobManager(conf *LambdaJobManagerConfig) (*lambdaJobManager, error) {

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	lambdaSvc := lambda.New(sess)

	return &lambdaJobManager{
		lambdaService: lambdaSvc,
		conf:          conf,
	}, nil
}

func (j *lambdaJobManager) Config() Config {
	return j.conf
}

func (j *lambdaJobManager) Launch(ctx context.Context, name string, spec Spec) error {

	if spec.Type() == FunctionJobType {
		functionSpec, ok := spec.(*FunctionSpec)
		if !ok {
			return ErrInvalidJobSpec
		}

		functionSpec.FunctionExtractPath = defaultLambdaFunctionExtractPath
	}
	storageConfig, err := spec.GetStorageConfig()
	if err != nil {
		return errors.Wrap(err, "Spec unexpectedly has no storage config.")
	}
	storageConfig.S3Config.AWSAccessKeyID = j.conf.AwsAccessKeyId
	storageConfig.S3Config.AWSSecretAccessKey = j.conf.AwsSecretAccessKey

	// Encode job spec to prevent data loss
	serializationType := JsonSerializationType
	encodedSpec, err := EncodeSpec(spec, serializationType)
	if err != nil {
		return err
	}

	lambdaFunctionRequest := map[string]string{"Spec": encodedSpec}
	payload, err := json.Marshal(lambdaFunctionRequest)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal request payload.")
	}

	functionName, err := mapJobTypeToLambdaFunction(spec)
	if err != nil {
		return errors.Wrap(err, "Unable to launch job.")
	}

	invokeInput := &lambda.InvokeInput{
		FunctionName:   &functionName,
		InvocationType: aws.String("Event"),
		Payload:        payload,
	}

	_, err = j.lambdaService.Invoke(invokeInput)
	if err != nil {
		return errors.Wrap(err, "Unable to invoke lambda function.")
	}
	return nil

}

func (j *lambdaJobManager) Poll(ctx context.Context, name string) (shared.ExecutionStatus, error) {
	return shared.UnknownExecutionStatus, ErrAsyncExecution
}

func (j *lambdaJobManager) DeployCronJob(ctx context.Context, name string, period string, spec Spec) error {
	return nil
}

func (j *lambdaJobManager) CronJobExists(ctx context.Context, name string) bool {
	return false
}

func (j *lambdaJobManager) EditCronJob(ctx context.Context, name string, cronString string) error {
	return nil
}

func (j *lambdaJobManager) DeleteCronJob(ctx context.Context, name string) error {
	return nil
}

// Maps a job Spec to Docker image.
func mapJobTypeToLambdaFunction(spec Spec) (string, error) {
	switch spec.Type() {
	case FunctionJobType:
		return lambda_utils.FunctionLambdaFunction, nil
	case AuthenticateJobType:
		authenticateSpec := spec.(*AuthenticateSpec)
		return mapIntegrationServiceToLambdaFunction(authenticateSpec.ConnectorName)
	case ExtractJobType:
		extractSpec := spec.(*ExtractSpec)
		return mapIntegrationServiceToLambdaFunction(extractSpec.ConnectorName)
	case LoadJobType:
		loadSpec := spec.(*LoadSpec)
		return mapIntegrationServiceToLambdaFunction(loadSpec.ConnectorName)
	case DiscoverJobType:
		discoverSpec := spec.(*DiscoverSpec)
		return mapIntegrationServiceToLambdaFunction(discoverSpec.ConnectorName)
	case ParamJobType:
		return lambda_utils.ParameterLambdaFunction, nil
	case SystemMetricJobType:
		return lambda_utils.SystemMetricLambdaFunction, nil
	default:
		return "", errors.Newf("Unsupported job type %v provided", spec.Type())
	}
}

func mapIntegrationServiceToLambdaFunction(service integration.Service) (string, error) {
	switch service {
	case integration.Snowflake:
		return lambda_utils.SnowflakeLambdaFunction, nil
	case integration.Postgres, integration.Redshift, integration.AqueductDemo:
		return lambda_utils.PostgresLambdaFunction, nil
	case integration.BigQuery:
		return lambda_utils.BigQueryLambdaFunction, nil
	case integration.S3:
		return lambda_utils.S3LambdaFunction, nil
	case integration.Athena:
		return lambda_utils.AthenaLambdaFunction, nil
	default:
		return "", errors.Newf("Unknown integration service provided %v", service)
	}
}
