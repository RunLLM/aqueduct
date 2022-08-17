package k8s

const (
	// The namespaces in which we create Kubernetes pods. As is obvious, the
	// `UserNamespace` is where user workload pods will be deployed, and the
	// `SystemNamespace` is where system services will be deployed.
	// `KubeSystemNamespace` is where kubernetes system resources will be deployed.
	UserNamespace = "user"

	DefaultCPURequest    = "2"
	DefaultMemoryRequest = "4Gi"

	// Pod Config
	PodResourceCPUKey       = "cpu"
	PodResourceMemoryKey    = "memory"
	PodSelectorLabelRoleKey = "role"

	// Cluster constants
	DockerSecretName = "regcred"
	DockerServer     = "https://index.docker.io/v2/" // this corresponds to the Docker Hub server address

	// The name of the k8s secret for the AWS credentials.
	AwsCredentialsSecretName = "awscred"
	AwsAccessKeyIdName       = "AWS_ACCESS_KEY_ID"
	AwsAccessKeyName         = "AWS_SECRET_ACCESS_KEY"
)
