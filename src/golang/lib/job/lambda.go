package job

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/operator/function"
	lambda_utils "github.com/aqueducthq/aqueduct/lib/lambda"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
)

const (
	defaultLambdaFunctionExtractPath = "/tmp/app/function/"
	updateFunctionMemoryTimeout      = 2 * time.Minute
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

// Updates the amount of memory available to the give function. Returns the previous memory setting, in MB.
func (j *lambdaJobManager) updateFunctionMemory(
	ctx context.Context,
	functionName string,
	newMemoryMB *int64,
) (*int64, error) {
	prevLambdaFnConfig, err := j.lambdaService.GetFunctionConfigurationWithContext(
		ctx,
		&lambda.GetFunctionConfigurationInput{
			FunctionName: &functionName,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to query Lambda to configure custom memory.")
	}

	prevMemoryMB := prevLambdaFnConfig.MemorySize

	latestLambdaFnConfig, err := j.lambdaService.UpdateFunctionConfigurationWithContext(
		ctx,
		&lambda.UpdateFunctionConfigurationInput{
			FunctionName: &functionName,
			MemorySize:   newMemoryMB,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to update Lambda function with custom memory.")
	}

	// Wait for at most a few minutes for the configuration to update.
	start := time.Now()
	for {
		// Check if memory has been updated yet.
		if *latestLambdaFnConfig.LastUpdateStatus == lambda.LastUpdateStatusSuccessful &&
			*latestLambdaFnConfig.MemorySize == *newMemoryMB {
			break
		} else if *latestLambdaFnConfig.LastUpdateStatus == lambda.LastUpdateStatusFailed {
			return nil, errors.Newf(
				"Unable to update Lambda with custom memory: %v",
				*latestLambdaFnConfig.LastUpdateStatusReason,
			)
		}

		polledLambdaFnConfig, err := j.lambdaService.GetFunctionConfigurationWithContext(
			ctx,
			&lambda.GetFunctionConfigurationInput{
				FunctionName: &functionName,
			},
		)
		if err != nil {
			// Ignore the error, keep polling until we hit the timeout.
			log.Errorf("Error when polling Lambda for function configuration: %v", err)
			continue
		}

		latestLambdaFnConfig = polledLambdaFnConfig

		if time.Since(start) > updateFunctionMemoryTimeout {
			return nil, errors.New("Unable to update Lambda function with custom memory. The operator timed out.")
		}
		time.Sleep(2 * time.Second)
	}

	return prevMemoryMB, nil
}

func (j *lambdaJobManager) Launch(ctx context.Context, name string, spec Spec) JobError {
	functionName, err := mapJobTypeToLambdaFunction(spec)
	if err != nil {
		return systemError(err)
	}

	// If set, we'll need to reset the function's memory back to this value after invocation.
	var previousMemoryMB *int64

	if spec.Type() == FunctionJobType {
		functionSpec, ok := spec.(*FunctionSpec)
		if !ok {
			return systemError(errors.Newf("Expected FunctionSpec, got %v", spec))
		}

		functionSpec.FunctionExtractPath = defaultLambdaFunctionExtractPath

		if functionSpec.Resources != nil {
			if functionSpec.Resources.MemoryMB != nil {
				// Resetting memory back to its original value is best-effort.
				defer func() {
					if previousMemoryMB != nil {
						_, err = j.updateFunctionMemory(ctx, functionName, previousMemoryMB)
						if err != nil {
							log.Errorf("Unable to reset function memory back to %v MB: %v", previousMemoryMB, err)
						}
					}
				}()

				newMemoryMBInt64 := int64(*functionSpec.Resources.MemoryMB)
				previousMemoryMB, err = j.updateFunctionMemory(ctx, functionName, &newMemoryMBInt64)
				if err != nil {
					return systemError(err)
				}
			}
		}
	}

	storageConfig, err := spec.GetStorageConfig()
	if err != nil {
		return systemError(errors.Wrap(err, "Spec unexpectedly has no storage config."))
	}
	storageConfig.S3Config.AWSAccessKeyID = j.conf.AwsAccessKeyId
	storageConfig.S3Config.AWSSecretAccessKey = j.conf.AwsSecretAccessKey

	// Encode job spec to prevent data loss
	serializationType := JsonSerializationType
	encodedSpec, err := EncodeSpec(spec, serializationType)
	if err != nil {
		return systemError(err)
	}

	lambdaFunctionRequest := map[string]string{"Spec": encodedSpec}
	payload, err := json.Marshal(lambdaFunctionRequest)
	if err != nil {
		return systemError(errors.Wrap(err, "Unable to marshal request payload."))
	}

	// Lambda functions with custom memory configurations should be executed synchronously.
	// This is to prevent such memory configurations from bleeding out to other operators using
	// the same lambda function. We reset the memory configuration at the very end of this function.
	// NOTE: this does not provide perfect isolation. It is still possible for operators scheduled
	// before to race with the memory configuration update, since regular lambda functions are run
	// asynchronously, and we have no visibility into the AWS's event queue.
	invocationType := aws.String("Event")
	if previousMemoryMB != nil {
		invocationType = aws.String("RequestResponse")
	}

	invokeInput := &lambda.InvokeInput{
		FunctionName:   &functionName,
		InvocationType: invocationType,
		Payload:        payload,
	}

	_, err = j.lambdaService.InvokeWithContext(ctx, invokeInput)
	if err != nil {
		return systemError(errors.Wrap(err, "Unable to invoke lambda function."))
	}
	return nil
}

func (j *lambdaJobManager) Poll(ctx context.Context, name string) (shared.ExecutionStatus, JobError) {
	return shared.UnknownExecutionStatus, noopError(errors.New("Cannot poll a lambda job manager."))
}

func (j *lambdaJobManager) DeployCronJob(ctx context.Context, name string, period string, spec Spec) JobError {
	return nil
}

func (j *lambdaJobManager) CronJobExists(ctx context.Context, name string) bool {
	return false
}

func (j *lambdaJobManager) EditCronJob(ctx context.Context, name string, cronString string) JobError {
	return nil
}

func (j *lambdaJobManager) DeleteCronJob(ctx context.Context, name string) JobError {
	return nil
}

// Maps a job Spec to Docker image.
func mapJobTypeToLambdaFunction(spec Spec) (string, error) {
	switch spec.Type() {
	case FunctionJobType:
		functionSpec, ok := spec.(*FunctionSpec)
		if !ok {
			return "", errors.New("Unable to determine Python Version.")
		}
		pythonVersion, err := function.GetPythonVersion(context.TODO(), functionSpec.FunctionPath, &functionSpec.StorageConfig)
		if err != nil {
			return "", errors.New("Unable to determine Python Version.")
		}
		switch pythonVersion {
		case function.PythonVersion37:
			return lambda_utils.FunctionLambdaFunction37, nil
		case function.PythonVersion38:
			return lambda_utils.FunctionLambdaFunction38, nil
		case function.PythonVersion39:
			return lambda_utils.FunctionLambdaFunction39, nil
		case function.PythonVersion310:
			return "", errors.New("Lambda does not support Python 3.10")
		default:
			return "", errors.New("Unable to determine Python Version.")
		}
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

func mapIntegrationServiceToLambdaFunction(service shared.Service) (string, error) {
	switch service {
	case shared.Snowflake:
		return lambda_utils.SnowflakeLambdaFunction, nil
	case shared.Postgres, shared.Redshift, shared.AqueductDemo:
		return lambda_utils.PostgresLambdaFunction, nil
	case shared.BigQuery:
		return lambda_utils.BigQueryLambdaFunction, nil
	case shared.S3:
		return lambda_utils.S3LambdaFunction, nil
	case shared.Athena:
		return lambda_utils.AthenaLambdaFunction, nil
	default:
		return "", errors.Newf("Unknown integration service provided %v", service)
	}
}
