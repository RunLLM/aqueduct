package job

import (
	"encoding/gob"
)

type ManagerType string

const (
	ProcessType ManagerType = "process"
	K8sType     ManagerType = "k8s"
)

type Config interface {
	Type() ManagerType
}

type ProcessConfig struct {
	BinaryDir             string `yaml:"binaryDir" json:"binary_dir"`
	LogsDir               string `yaml:"logsDir" json:"logs_dir"`
	PythonExecutorPackage string `yaml:"pythonExecutorPackage" json:"python_executor_package"`
	OperatorStorageDir    string `yaml:"operatorStorageDir" json:"operator_storage_dir"`
}

type K8sJobManagerConfig struct {
	KubeconfigPath     string `yaml:"kubeconfigPath" json:"kube_config_path"`
	ClusterName        string `yaml:"clusterName" json:"cluster_name"`
	AwsAccessKeyId     string `yaml:"awsAccessKeyId" json:"aws_access_key_id"`
	AwsSecretAccessKey string `yaml:"awsSecretAccessKey" json:"aws_secret_access_key"`

	// System config, will have defaults
	AwsRegion                        string `yaml:"awsRegion" json:"aws_region"`
	FunctionDockerImage              string `yaml:"functionDockerImage" json:"function_docker_image"`
	ParameterDockerImage             string `yaml:"parameterDockerImage" json:"parameter_docker_image"`
	PostgresConnectorDockerImage     string `yaml:"postgresConnectorDockerImage" json:"postgres_connector_docker_image"`
	SnowflakeConnectorDockerImage    string `yaml:"snowflakeConnectorDockerImage" json:"snowflake_connector_docker_image"`
	MySqlConnectorDockerImage        string `yaml:"mySqlConnectorDockerImage" json:"mySql_connector_docker_image"`
	SqlServerConnectorDockerImage    string `yaml:"sqlServerConnectorDockerImage" json:"sql_server_connector_docker_image"`
	BigQueryConnectorDockerImage     string `yaml:"bigQueryConnectorDockerImage" json:"big_query_connector_docker_image"`
	GoogleSheetsConnectorDockerImage string `yaml:"googleSheetsConnectorDockerImage" json:"google_sheets_connector_docker_image"`
	SalesforceConnectorDockerImage   string `yaml:"salesforceConnectorDockerImage" json:"salesforce_connector_docker_image"`
	S3ConnectorDockerImage           string `yaml:"s3ConnectorDockerImage" json:"s3_connector_docker_image"`
}

func (*K8sJobManagerConfig) Type() ManagerType {
	return K8sType
}

func (*ProcessConfig) Type() ManagerType {
	return ProcessType
}

func RegisterGobTypes() {
	gob.Register(&ProcessConfig{})
	gob.Register(&K8sJobManagerConfig{})
	gob.Register(&WorkflowSpec{})
	gob.Register(&WorkflowRetentionSpec{})
}

func init() {
	RegisterGobTypes()
}
