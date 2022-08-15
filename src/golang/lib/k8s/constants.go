package k8s

const (
	// The namespaces in which we create Kubernetes pods. As is obvious, the
	// `UserNamespace` is where user workload pods will be deployed, and the
	// `SystemNamespace` is where system services will be deployed.
	// `KubeSystemNamespace` is where kubernetes system resources will be deployed.
	UserNamespace       = "user"
	SystemNamespace     = "spiral-system"
	DefaultNamespace    = "default"
	KubeSystemNamespace = "kube-system"

	DefaultCPURequest    = "2"
	DefaultMemoryRequest = "4Gi"

	ServiceCreationTimeoutSec = 60

	// Pod Config
	PodResourceCPUKey       = "cpu"
	PodResourceMemoryKey    = "memory"
	PodSelectorLabelRoleKey = "role"

	// Healthiness

	// Pod Probes:

	// Use this port to probe if all ports are open
	DefaultProbingPort = 2333
	// The following probe determines if the container is successfully initialized. It tries to
	// establish connection to the container.
	//
	// The numbers in configs means:
	// Probe starts after 10 secs the container starts. It tries to establish connection
	// to the probing port every 5s, and once the connection establishes, it's considered
	// as success and the container is considered as successfully started. Otherwise, the
	// probe fails after 60 tries, or 300 seconds.
	StartupProbeTimeoutSec       = 300
	StartupProbeInitialDelaySec  = 10
	StartupProbePeriodSec        = 5
	StartupProbeSuccessThreshold = 1
	StartupProbeFailureThreshold = 60

	// The following probe determines if the running container is ready. It tries to
	// establish connection to the container.
	//
	// The numbers in configs means:
	// Probe starts after 1 secs after the above probe succeeded. It tries to establish connection
	// to the probing port every 20s, and once the connection establishes, it's considered
	// as success and the container is considered as successfully started.
	//
	// The probe is considered as failure either
	// - never success after the first 60s and timedout
	// - 2 consecutive failures after a success
	//
	// When the probe failed, it will be removed from the service endpoint
	ReadinessProbeInitialDelaySec  = 1
	ReadinessProbeTimeoutSec       = 60
	ReadinessProbePeriodSec        = 20
	ReadinessProbeSuccessThreshold = 1
	ReadinessProbeFailureThreshold = 2

	// Horizontal Pod Autoscaler (HPA)
	// The following configs defines the scaling behavior that:
	// - Take action based on metrics collected in past 20 secs
	// - Scale up/down if the avg CPU across all pods is above / below 70%
	// - Avoid changing more than 3 pods in a window of 120 secs
	HpaPodChangeLimit              = 3
	HpaPodChangeLimitPeriod        = 120
	TargetCPUUtilizationPercentage = 70
	StablizationWindowSec          = 20
	// TODO: maybe make the replica upperbound different per deployments
	MinReplicas = 1
	MaxReplicas = 10

	CheckDeploymentAvailabilityTimeoutSec = 300

	CheckIngressAddressTimeoutSec = 300

	GRPCServerPingWaitEnforcementThreshold = 10
	GRPCClientPingThreshold                = 20
	GRPCClientPingTimeout                  = 20

	// A string template for the name of the Kubernetes role.
	ServiceRoleTemplate        = "%s-role"
	ServiceClusterRoleTemplate = "%s-cluster-role"

	// A string template for the name of the Kubernetes role binding that applies
	// the role created from the above name template to the serviceaccount.
	ServiceRoleBindingTemplate        = "%s-role-binding"
	ServiceClusterRoleBindingTemplate = "%s-cluster-role-binding"

	ExecutorServiceAccount = "executor-service"
	FunctionServiceAccount = "function-service"

	// Cluster constants
	DockerSecretName = "regcred"
	DockerServer     = "https://index.docker.io/v2/" // this corresponds to the Docker Hub server address

	// Image puller related constants
	ImagePullerDaemonsetName   = "imagepuller"
	ImagePullerDeployScript    = "config/deploy_image_puller.sh"
	ImagePullerManifest        = "config/image-puller.yml"
	UpdatedImagePullerManifest = "config/.image-puller-updated.yml"

	// The name of the k8s secret for the AWS credentials.
	AwsCredentialsSecretName = "awscred"

	// The ARN of the default AWS policy that gives a role access to all S3 buckets.
	AwsS3AccessArn = "arn:aws:iam::aws:policy/AmazonS3FullAccess"

	// The name of the role we create to give Function pods S3 access.
	AwsFunctionRoleName = "EKSFunctionS3Access"

	// This is a trust relationship policy document to attach an IAM role
	// to the specified service account.
	AwsRoleTrustRelationship = `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
						"Federated": "%s"
				},
				"Action": "sts:AssumeRoleWithWebIdentity",
				"Condition": {
					"StringEquals": {
						"%s:sub": "system:serviceaccount:%s:%s"
					}
				}
			}
		]
	}`
)
