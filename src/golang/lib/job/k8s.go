package job

import (
	"context"
	"os/exec"

	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/k8s"
	"github.com/dropbox/godropbox/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

const (
	defaultFunctionExtractPath = "/app/function/"
	jobSpecEnvVarKey           = "JOB_SPEC"
	DevBranchKey               = "PULL_BRANCH"
	ClusterEnvironmentKey      = "CLUSTER_ENVIRONMENT"
)

type k8sJobManager struct {
	k8sClient *kubernetes.Clientset
	conf      *K8sConfig
}

func NewK8sJobManager(conf *K8sConfig) (*k8sJobManager, error) {
	k8sClient, err := k8s.CreateClientOutsideCluster(conf.KubeConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "Error while creating K8sJobManager")
	}
	// create function service account
	// add docker secrets + image pulling
	// add aws credentials as secrets??? <- might not need if we are adding the service accounts

	// Update kubeconfig file
	cmd := exec.Command(
		"aws",
		"eks",
		"update-kubeconfig",
		"--region", conf.AwsRegion,
		"--name", conf.ClusterName,
		"--kubeconfig", conf.KubeConfigPath,
	)
	err = cmd.Run()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to update kubeconfig.")
	}

	k8s.CreateNamespaces(k8sClient)
	k8s.CreateAwsCredentialsSecret(conf.AwsAccessKeyId, conf.AwsSecretAccessKey, conf.KubeConfigPath)

	return &k8sJobManager{
		k8sClient: k8sClient,
		conf:      conf,
	}, nil
}

func (j *k8sJobManager) Config() Config {
	return j.conf
}
func (j *k8sJobManager) Launch(ctx context.Context, name string, spec Spec) error {
	logrus.Info("In K8sJobManager.Launch()")
	logrus.Info(spec)
	resourceRequest := generateResourceRequest(j.conf, spec.Type())
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

	// TODO: https://linear.app/aqueducthq/issue/ENG-369/create-k8s-service-accounts-for-local-minikube-clusters
	var secretEnvVars []string
	secretEnvVars = []string{k8s.AwsCredentialsSecretName}

	containerImage, err := mapJobTypeToDockerImage(j, spec)
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
func mapJobTypeToDockerImage(j *k8sJobManager, spec Spec) (string, error) {
	switch spec.Type() {
	// case WorkflowJobType:
	// 	return j.conf.ExecutorDockerImage, nil
	case FunctionJobType:
		return j.conf.FunctionDockerImage, nil
	case AuthenticateJobType:
		authenticateSpec := spec.(*AuthenticateSpec)
		return mapIntegrationServiceToDockerImage(j, authenticateSpec.ConnectorName)
	case ExtractJobType:
		extractSpec := spec.(*ExtractSpec)
		return mapIntegrationServiceToDockerImage(j, extractSpec.ConnectorName)
	case LoadJobType:
		loadSpec := spec.(*LoadSpec)
		return mapIntegrationServiceToDockerImage(j, loadSpec.ConnectorName)
	case DiscoverJobType:
		discoverSpec := spec.(*DiscoverSpec)
		return mapIntegrationServiceToDockerImage(j, discoverSpec.ConnectorName)
	case ParamJobType:
		return j.conf.ParameterDockerImage, nil
	default:
		return "", errors.Newf("Unsupported job type %v provided", spec.Type())
	}
}

func mapIntegrationServiceToDockerImage(j *k8sJobManager, service integration.Service) (string, error) {
	switch service {
	case integration.Postgres, integration.Redshift, integration.AqueductDemo:
		return j.conf.PostgresConnectorDockerImage, nil
	case integration.Snowflake:
		return j.conf.SnowflakeConnectorDockerImage, nil
	case integration.MySql, integration.MariaDb:
		return j.conf.MySqlConnectorDockerImage, nil
	case integration.SqlServer:
		return j.conf.SqlServerConnectorDockerImage, nil
	case integration.BigQuery:
		return j.conf.BigQueryConnectorDockerImage, nil
	case integration.GoogleSheets:
		return j.conf.GoogleSheetsConnectorDockerImage, nil
	case integration.Salesforce:
		return j.conf.SalesforceConnectorDockerImage, nil
	case integration.S3:
		return j.conf.S3ConnectorDockerImage, nil
	default:
		return "", errors.Newf("Unknown integration service provided %v", service)
	}
}

func generateResourceRequest(conf *K8sConfig, jobType JobType) map[string]string {
	resourceRequest := map[string]string{
		k8s.PodResourceCPUKey:    k8s.DefaultCPURequest,
		k8s.PodResourceMemoryKey: k8s.DefaultMemoryRequest,
	}

	return resourceRequest
}

// generateS3Annotation generates an annotation to be attached to the service account to allow
// it to access S3.
func generateS3Annotation(
	serviceAccount string,
	namespace string,
	roleName string,
	oidcIssuerUri *string,
	openIDConnectProviderArn string,
	awsRegion string,
	clusterName string,
) map[string]string {
	arn := k8s.CreateAwsFullS3Role(
		serviceAccount,
		namespace,
		roleName,
		oidcIssuerUri,
		openIDConnectProviderArn,
		awsRegion,
		clusterName,
	)

	return map[string]string{
		"eks.amazonaws.com/role-arn": arn,
	}
}
