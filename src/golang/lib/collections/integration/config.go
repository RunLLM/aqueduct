package integration

import (
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/dropbox/godropbox/errors"
)

type S3ConfigType string

const (
	AccessKeyS3ConfigType         S3ConfigType = "access_key"
	ConfigFilePathS3ConfigType    S3ConfigType = "config_file_path"
	ConfigFileContentS3ConfigType S3ConfigType = "config_file_content"
)

// S3Config contains the fields for connecting an S3 integration.
type S3Config struct {
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

// AirflowConfig contains the fields for connecting an Airflow integration.
type AirflowConfig struct{}

// GCSConfig contains the fields for connecting a Google Cloud Storage integration.
type GCSConfig struct {
	shared.GCSConfig
	UseAsStorage ConfigBool `json:"use_as_storage"`
}

type K8sIntegrationConfig struct {
	KubeconfigPath string     `json:"kubeconfig_path" yaml:"kubeconfigPath"`
	ClusterName    string     `json:"cluster_name"  yaml:"clusterName"`
	UseSameCluster ConfigBool `json:"use_same_cluster"  yaml:"useSameCluster"`
}

type LambdaIntegrationConfig struct {
	RoleArn string `json:"role_arn" yaml:"roleArn"`
}

type DatabricksIntegrationConfig struct {
	WorkspaceUrl         string `json:"workspace_url" yaml:"workspaceUrl"`
	AccessToken          string `json:"access_token" yaml:"accessToken"`
	S3InstanceProfileArn string `json:"s3_instance_profile_arn" yaml:"s3InstanceProfileArn"`
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
