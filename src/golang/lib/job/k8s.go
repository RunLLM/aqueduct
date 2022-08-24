package job

import (
	"context"
	"io/ioutil"

	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/k8s"
	"github.com/dropbox/godropbox/errors"
	"k8s.io/client-go/kubernetes"
)

const (
	defaultFunctionExtractPath = "/app/function/"
	jobSpecEnvVarKey           = "JOB_SPEC"
	gcsCredentialsEnvVarKey    = "GCS_CREDENTIALS"
)

type k8sJobManager struct {
	k8sClient *kubernetes.Clientset
	conf      *K8sJobManagerConfig
}

func NewK8sJobManager(conf *K8sJobManagerConfig) (*k8sJobManager, error) {
	k8sClient, err := k8s.CreateClientOutsideCluster(conf.KubeconfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "Error while creating K8sJobManager")
	}

	err = k8s.CreateNamespaces(k8sClient)
	if err != nil {
		return nil, errors.Wrap(err, "Error while creating K8sJobManager")
	}

	secretsMap := map[string]string{}
	secretsMap[k8s.AwsAccessKeyIdName] = conf.AwsAccessKeyId
	secretsMap[k8s.AwsAccessKeyName] = conf.AwsSecretAccessKey
	err = k8s.CreateSecret(context.TODO(), k8s.AwsCredentialsSecretName, secretsMap, k8sClient)
	if err != nil {
		return nil, errors.Wrap(err, "Error while creating K8sJobManager")
	}

	return &k8sJobManager{
		k8sClient: k8sClient,
		conf:      conf,
	}, nil
}

func (j *k8sJobManager) Config() Config {
	return j.conf
}

func (j *k8sJobManager) Launch(ctx context.Context, name string, spec Spec) error {
	resourceRequest := map[string]string{
		k8s.PodResourceCPUKey:    k8s.DefaultCPURequest,
		k8s.PodResourceMemoryKey: k8s.DefaultMemoryRequest,
	}
	environmentVariables := map[string]string{}

	if spec.Type() == FunctionJobType {
		functionSpec, ok := spec.(*FunctionSpec)
		if !ok {
			return ErrInvalidJobSpec
		}

		functionSpec.FunctionExtractPath = defaultFunctionExtractPath
	}

	// Encode job spec to prevent data loss
	serializationType := JsonSerializationType
	encodedSpec, err := EncodeSpec(spec, serializationType)
	if err != nil {
		return err
	}

	environmentVariables[jobSpecEnvVarKey] = encodedSpec

	secretEnvVars := []string{}

	if spec.HasStorageConfig() {
		storageConfig, err := spec.GetStorageConfig()
		if err != nil {
			return err
		}

		switch storageConfig.Type {
		case shared.S3StorageType:
			// k8s clusters access S3 via credentials passed as a secret
			secretEnvVars = append(secretEnvVars, k8s.AwsCredentialsSecretName)
		case shared.GCSStorageType:
			// For GCS the credentials must be provided as an environment variable
			data, err := ioutil.ReadFile(storageConfig.GCSConfig.CredentialsPath)
			if err != nil {
				return err
			}
			environmentVariables[gcsCredentialsEnvVarKey] = string(data)
		default:
			return errors.Newf("Storage type %v is not supported for k8s job managers", storageConfig.Type)
		}
	}

	containerImage, err := mapJobTypeToDockerImage(spec)
	if err != nil {
		return err
	}

	return k8s.LaunchJob(
		name,
		containerImage,
		&environmentVariables,
		secretEnvVars,
		&resourceRequest,
		j.k8sClient,
	)
}

func (j *k8sJobManager) Poll(ctx context.Context, name string) (shared.ExecutionStatus, error) {
	job, err := k8s.GetJob(name, j.k8sClient)
	if err != nil {
		return shared.UnknownExecutionStatus, ErrJobNotExist
	}

	var status shared.ExecutionStatus

	if job.Status.Succeeded == 1 {
		status = shared.SucceededExecutionStatus
	} else if job.Status.Failed == 1 {
		status = shared.FailedExecutionStatus
	} else {
		status = shared.PendingExecutionStatus
	}

	return status, nil
}

func (j *k8sJobManager) DeployCronJob(ctx context.Context, name string, period string, spec Spec) error {
	return nil
}

func (j *k8sJobManager) CronJobExists(ctx context.Context, name string) bool {
	return false
}

func (j *k8sJobManager) EditCronJob(ctx context.Context, name string, cronString string) error {
	return nil
}

func (j *k8sJobManager) DeleteCronJob(ctx context.Context, name string) error {
	return nil
}

// Maps a job Spec to Docker image.
func mapJobTypeToDockerImage(spec Spec) (string, error) {
	switch spec.Type() {
	// case WorkflowJobType:
	// 	return j.conf.ExecutorDockerImage, nil
	case FunctionJobType:
		return DefaultFunctionDockerImage, nil
	case AuthenticateJobType:
		authenticateSpec := spec.(*AuthenticateSpec)
		return mapIntegrationServiceToDockerImage(authenticateSpec.ConnectorName)
	case ExtractJobType:
		extractSpec := spec.(*ExtractSpec)
		return mapIntegrationServiceToDockerImage(extractSpec.ConnectorName)
	case LoadJobType:
		loadSpec := spec.(*LoadSpec)
		return mapIntegrationServiceToDockerImage(loadSpec.ConnectorName)
	case DiscoverJobType:
		discoverSpec := spec.(*DiscoverSpec)
		return mapIntegrationServiceToDockerImage(discoverSpec.ConnectorName)
	case ParamJobType:
		return DefaultParameterDockerImage, nil
	case SystemMetricJobType:
		return DefaultSystemMetricDockerImage, nil
	default:
		return "", errors.Newf("Unsupported job type %v provided", spec.Type())
	}
}

func mapIntegrationServiceToDockerImage(service integration.Service) (string, error) {
	switch service {
	case integration.Postgres, integration.Redshift, integration.AqueductDemo:
		return DefaultPostgresConnectorDockerImage, nil
	case integration.Snowflake:
		return DefaultSnowflakeConnectorDockerImage, nil
	case integration.MySql, integration.MariaDb:
		return DefaultMySqlConnectorDockerImage, nil
	case integration.SqlServer:
		return DefaultSqlServerConnectorDockerImage, nil
	case integration.BigQuery:
		return DefaultBigQueryConnectorDockerImage, nil
	case integration.S3:
		return DefaultS3ConnectorDockerImage, nil
	default:
		return "", errors.Newf("Unknown integration service provided %v", service)
	}
}
