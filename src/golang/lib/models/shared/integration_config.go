package shared

import (
	"fmt"

	"github.com/dropbox/godropbox/errors"
)

type S3ConfigType string

const (
	AccessKeyS3ConfigType         S3ConfigType = "access_key"
	ConfigFilePathS3ConfigType    S3ConfigType = "config_file_path"
	ConfigFileContentS3ConfigType S3ConfigType = "config_file_content"
)

// S3IntegrationConfig contains the fields for connecting an S3 integration.
type S3IntegrationConfig struct {
	Type              S3ConfigType `json:"type"`
	Bucket            string       `json:"bucket"`
	Region            string       `json:"region"`
	AccessKeyId       string       `json:"access_key_id"`
	SecretAccessKey   string       `json:"secret_access_key"`
	ConfigFilePath    string       `json:"config_file_path"`
	ConfigFileContent string       `json:"config_file_content"`
	ConfigFileProfile string       `json:"config_file_profile"`
	UseAsStorage      ConfigBool   `json:"use_as_storage"`
}

// AirflowIntegrationConfig contains the fields for connecting an Airflow integration.
type AirflowIntegrationConfig struct{}

// GCSIntegrationConfig contains the fields for connecting a Google Cloud Storage integration.
type GCSIntegrationConfig struct {
	GCSConfig
	UseAsStorage ConfigBool `json:"use_as_storage"`
}

type K8sClusterStatusType string

const (
	K8sClusterCreatingStatus    K8sClusterStatusType = "Creating"
	K8sClusterActiveStatus      K8sClusterStatusType = "Active"
	K8sClusterTerminatingStatus K8sClusterStatusType = "Terminating"
	K8sClusterTerminatedStatus  K8sClusterStatusType = "Terminated"

	K8sKubeconfigPathKey     string = "kubeconfig_path"
	K8sClusterNameKey        string = "cluster_name"
	K8sDynamicKey            string = "dynamic"
	K8sCloudIntegrationIdKey string = "cloud_integration_id"
	K8sUseSameClusterKey     string = "use_same_cluster"
	K8sStatusKey             string = "status"
	K8sKeepaliveKey          string = "keepalive"
	K8sLastUsedTimestampKey  string = "last_used_timestamp"
	DynamicK8sClusterName    string = "aqueduct_k8s"
	DefaultKeepalive         int    = 1200
)

type K8sIntegrationConfig struct {
	KubeconfigPath     string     `json:"kubeconfig_path" yaml:"kubeconfigPath"`
	ClusterName        string     `json:"cluster_name"  yaml:"clusterName"`
	UseSameCluster     ConfigBool `json:"use_same_cluster"  yaml:"useSameCluster"`
	Dynamic            ConfigBool `json:"dynamic"  yaml:"dynamic"`
	CloudIntegrationId string     `json:"cloud_integration_id"  yaml:"cloud_integration_id"`
}

type LambdaIntegrationConfig struct {
	RoleArn   string `json:"role_arn" yaml:"roleArn"`
	ExecState string `json:"exec_state" yaml:"execState"`
}

type DatabricksIntegrationConfig struct {
	// WorkspaceURL is the full url for the Databricks workspace that
	// Aqueduct operators will run on.
	WorkspaceURL string `json:"workspace_url" yaml:"workspaceUrl"`
	// AccessToken is a Databricks AccessToken for a workspace. Information on how
	// to create tokens can be found here: https://docs.databricks.com/dev-tools/auth.html#personal-access-tokens-for-users
	AccessToken string `json:"access_token" yaml:"accessToken"`
	// Databricks needs an Instance Profile with S3 permissions in order to access metadata
	// storage in S3. Information on how to create this can be found here:
	// https://docs.databricks.com/aws/iam/instance-profile-tutorial.html
	S3InstanceProfileARN string `json:"s3_instance_profile_arn" yaml:"s3InstanceProfileArn"`
}

type EmailConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	// Targets are email addresses for receivers.
	Targets []string          `json:"targets"`
	Level   NotificationLevel `json:"level"`
	Enabled bool              `json:"enabled"`
}

type SlackConfig struct {
	Token    string            `json:"token"`
	Channels []string          `json:"channels"`
	Level    NotificationLevel `json:"level"`
	Enabled  bool              `json:"enabled"`
}

type AWSConfig struct {
	AccessKeyId     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
}

type SparkIntegrationConfig struct {
	// LivyServerURL is the URL of the Livy server that sits in front of the Spark cluster.
	// This URL is assumed to be accessible by the machine running the Aqueduct server.
	LivyServerURL string `yaml:"baseUrl" json:"livy_server_url"`
	// AWS Access Key ID is passed from the StorageConfig.
	AwsAccessKeyID string `yaml:"awsAccessKeyId" json:"aws_access_key_id"`
	// AWS Secret Access Key is passed from the StorageConfig.
	AwsSecretAccessKey string `yaml:"awsSecretAccessKey" json:"aws_secret_access_key"`
}

func (c *EmailConfig) FullHost() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type ConfigBool bool

func (scb *ConfigBool) UnmarshalJSON(data []byte) error {
	s := string(data)
	var b bool

	// TODO ENG-1586: Remove hack of treating credential string as a boolean
	switch s {
	case "\"true\"":
		b = true
	case "\"false\"":
		b = false
	default:
		return errors.Newf("Unable to unmarshal %s into ConfigBool", s)
	}

	*scb = ConfigBool(b)
	return nil
}
