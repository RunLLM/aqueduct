package job

import (
	"context"
	"fmt"
	"github.com/aqueducthq/aqueduct/lib"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/operator/function"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/k8s"
	"github.com/dropbox/godropbox/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"strconv"
)

const (
	defaultFunctionExtractPath = "/app/function/"
	jobSpecEnvVarKey           = "JOB_SPEC"
)

type k8sJobManager struct {
	k8sClient *kubernetes.Clientset
	conf      *K8sJobManagerConfig
}

func NewK8sJobManager(conf *K8sJobManagerConfig) (*k8sJobManager, error) {
	k8sClient, err := k8s.CreateK8sClient(conf.KubeconfigPath, conf.UseSameCluster)
	if err != nil {
		return nil, errors.Wrap(err, "Error while creating K8sClient")
	}

	err = k8s.CreateNamespaces(k8sClient)
	if err != nil {
		return nil, errors.Wrap(err, "Error while creating K8s Namespaces")
	}

	secretsMap := map[string]string{}
	secretsMap[k8s.AwsAccessKeyIdName] = conf.AwsAccessKeyId
	secretsMap[k8s.AwsAccessKeyName] = conf.AwsSecretAccessKey
	err = k8s.CreateSecret(context.TODO(), k8s.AwsCredentialsSecretName, secretsMap, k8sClient)
	if err != nil {
		return nil, errors.Wrap(err, "Error while creating K8s Secrets")
	}

	return &k8sJobManager{
		k8sClient: k8sClient,
		conf:      conf,
	}, nil
}

func (j *k8sJobManager) Config() Config {
	return j.conf
}

func (j *k8sJobManager) Launch(ctx context.Context, name string, spec Spec) JobError {
	launchGpu := false
	resourceRequest := map[string]string{
		k8s.PodResourceCPUKey:    k8s.DefaultCPURequest,
		k8s.PodResourceMemoryKey: k8s.DefaultMemoryRequest,
	}

	environmentVariables := map[string]string{}

	if spec.Type() == FunctionJobType {
		functionSpec, ok := spec.(*FunctionSpec)
		if !ok {
			return systemError(errors.Newf("Function Spec is expected, but got %v", spec))
		}

		functionSpec.FunctionExtractPath = defaultFunctionExtractPath

		if functionSpec.Resources != nil {
			if functionSpec.Resources.GPUResourceName != nil {
				resourceRequest[k8s.GPUResourceName] = *functionSpec.Resources.GPUResourceName
				launchGpu = true
			}

			if functionSpec.Resources.NumCPU != nil {
				resourceRequest[k8s.PodResourceCPUKey] = strconv.Itoa(*functionSpec.Resources.NumCPU)
			}
			if functionSpec.Resources.MemoryMB != nil {
				// Set the request to be in "M" = Megabytes.
				resourceRequest[k8s.PodResourceMemoryKey] = fmt.Sprintf("%sM",
					strconv.Itoa(*functionSpec.Resources.MemoryMB),
				)
			}
		}
	}

	// Encode job spec to prevent data loss
	serializationType := JsonSerializationType
	encodedSpec, err := EncodeSpec(spec, serializationType)
	if err != nil {
		return systemError(err)
	}

	environmentVariables[jobSpecEnvVarKey] = encodedSpec

	secretEnvVars := []string{}

	if spec.HasStorageConfig() {
		// This job spec has a storage config that k8s needs access to
		storageConfig, err := spec.GetStorageConfig()
		if err != nil {
			return systemError(err)
		}

		if storageConfig.Type == shared.S3StorageType {
			// k8s clusters access S3 via credentials passed as a secret
			secretEnvVars = append(secretEnvVars, k8s.AwsCredentialsSecretName)
		}
	}

	containerRepo, err := mapJobTypeToDockerImage(spec, launchGpu)
	if err != nil {
		return systemError(err)
	}
	containerImage := fmt.Sprintf("%s:%s", containerRepo, lib.ServerVersionNumber)

	err = k8s.LaunchJob(
		name,
		containerImage,
		&environmentVariables,
		secretEnvVars,
		&resourceRequest,
		j.k8sClient,
	)
	if err != nil {
		return systemError(err)
	}
	return nil
}

func containerStatusFromPod(pod *corev1.Pod, name string) (*corev1.ContainerStatus, error) {
	if len(pod.Status.ContainerStatuses) != 1 {
		return nil, errors.Newf(
			"Expected job %s to have one container, but instead got %v.",
			name,
			len(pod.Status.ContainerStatuses),
		)
	}

	containerStatus := pod.Status.ContainerStatuses[0]
	if containerStatus.State.Terminated == nil {
		return nil, errors.Newf(
			"Container %s should have terminated.", containerStatus.Name,
		)
	}
	return &containerStatus, nil
}

func (j *k8sJobManager) Poll(ctx context.Context, name string) (shared.ExecutionStatus, JobError) {
	job, err := k8s.GetJob(ctx, name, j.k8sClient)
	if err != nil {
		return shared.UnknownExecutionStatus, jobMissingError(err)
	}

	var status shared.ExecutionStatus
	if job.Status.Succeeded == 1 {
		status = shared.SucceededExecutionStatus
	} else if job.Status.Failed == 1 {
		status = shared.FailedExecutionStatus

		// Fetch more detailed information about the failure, in case there is valuable
		// context we can surface to the user.
		pod, err := k8s.GetPod(ctx, name, j.k8sClient)
		if err != nil {
			return status, systemError(err)
		}

		containerStatus, err := containerStatusFromPod(pod, name)
		if err != nil {
			return status, systemError(err)
		}

		if containerStatus.State.Terminated.Reason == "OOMKilled" {
			return status, userError(errors.New("Operator failed on Kubernetes due to Out-of-Memory exception."))
		}

		// TODO: look at the other job managers...
		// We do not error here since pods are killed with a failing exit status on any failed checks.
		// We should rely on the written execution state to decide whether to continue dag execution,
		// and not the status of the pod.
		return status, nil
	} else {
		pod, err := k8s.GetPod(ctx, name, j.k8sClient)
		if err != nil {
			if err == errors.New("No pod exists") {
				return shared.PendingExecutionStatus, nil
			}
			return shared.FailedExecutionStatus, systemError(err)
		}

		if pod.Status.Phase == corev1.PodPending {
			for _, condition := range pod.Status.Conditions {
				// If the pod is unschedulable, fail the operator immediately with the K8s message (eg. "Insufficient CPU").
				if condition.Type == corev1.PodScheduled &&
					condition.Status == corev1.ConditionFalse &&
					condition.Reason == corev1.PodReasonUnschedulable {
					return shared.FailedExecutionStatus, userError(
						errors.Newf("Pod could not be scheduled on Kubernetes. Reason: %s", condition.Message),
					)
				}
			}
		}

		status = shared.PendingExecutionStatus
	}

	return status, nil
}

func (j *k8sJobManager) DeployCronJob(ctx context.Context, name string, period string, spec Spec) JobError {
	return nil
}

func (j *k8sJobManager) CronJobExists(ctx context.Context, name string) bool {
	return false
}

func (j *k8sJobManager) EditCronJob(ctx context.Context, name string, cronString string) JobError {
	return nil
}

func (j *k8sJobManager) DeleteCronJob(ctx context.Context, name string) JobError {
	return nil
}

// Maps a job Spec to Docker image.
func mapJobTypeToDockerImage(spec Spec, launchGpu bool) (string, error) {
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
		if launchGpu {
			switch pythonVersion {
			case function.PythonVersion37:
				return GpuFunction37DockerImage, nil
			case function.PythonVersion38:
				return GpuFunction38DockerImage, nil
			case function.PythonVersion39:
				return GpuFunction39DockerImage, nil
			case function.PythonVersion310:
				return GpuFunction310DockerImage, nil
			default:
				return "", errors.New("Unable to determine Python Version.")
			}
		} else {
			switch pythonVersion {
			case function.PythonVersion37:
				return Function37DockerImage, nil
			case function.PythonVersion38:
				return Function38DockerImage, nil
			case function.PythonVersion39:
				return Function39DockerImage, nil
			case function.PythonVersion310:
				return Function310DockerImage, nil
			default:
				return "", errors.New("Unable to determine Python Version.")
			}
		}

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
		return ParameterDockerImage, nil
	case SystemMetricJobType:
		return SystemMetricDockerImage, nil
	default:
		return "", errors.Newf("Unsupported job type %v provided", spec.Type())
	}
}

func mapIntegrationServiceToDockerImage(service integration.Service) (string, error) {
	switch service {
	case integration.Postgres, integration.Redshift, integration.AqueductDemo:
		return PostgresConnectorDockerImage, nil
	case integration.Snowflake:
		return SnowflakeConnectorDockerImage, nil
	case integration.MySql, integration.MariaDb:
		return MySqlConnectorDockerImage, nil
	case integration.SqlServer:
		return SqlServerConnectorDockerImage, nil
	case integration.BigQuery:
		return BigQueryConnectorDockerImage, nil
	case integration.S3:
		return S3ConnectorDockerImage, nil
	default:
		return "", errors.Newf("Unknown integration service provided %v", service)
	}
}
