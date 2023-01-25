package shared

import (
	"database/sql/driver"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
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
	AccessKeyID       string       `json:"access_key_id"`
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

type K8sIntegrationConfig struct {
	KubeconfigPath string     `json:"kubeconfig_path" yaml:"kubeconfigPath"`
	ClusterName    string     `json:"cluster_name"  yaml:"clusterName"`
	UseSameCluster ConfigBool `json:"use_same_cluster"  yaml:"useSameCluster"`
}

type EmailConfig struct {
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	Host     string `json:"host" yaml:"host"`
	Port     string `json:"port" yaml:"port"`
	// Targets are email addresses for receivers.
	Targets []string          `json:"targets" yaml:"targets"`
	Level   NotificationLevel `json:"level" yaml:"level"`
}

type SlackConfig struct {
	Token    string            `json:"token" yaml:"token"`
	Channels []string          `json:"channels" yaml:"channels"`
	Level    NotificationLevel `json:"level" yaml:"level"`
}

func (c *EmailConfig) FullHost() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type LambdaIntegrationConfig struct {
	RoleArn string `json:"role_arn" yaml:"roleArn"`
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

type IntegrationConfig map[string]string

func (c *IntegrationConfig) Value() (driver.Value, error) {
	return utils.ValueJSONB(*c)
}

func (c *IntegrationConfig) Scan(value interface{}) error {
	return utils.ScanJSONB(value, c)
}
