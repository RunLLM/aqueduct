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

type K8sConfig struct {
	MinikubeSystemPodCpu    string `yaml:"minikubeSystemPodCpu" json:"minikube_system_pod_cpu"`
	MinikubeSystemPodMemory string `yaml:"minikubeSystemPodMemory" json:"minikube_system_pod_memory"`
	MinikubeUserPodCpu      string `yaml:"minikubeUserPodCpu" json:"minikube_user_pod_cpu"`
	MinikubeUserPodMemory   string `yaml:"minikubeUserPodMemory" json:"minikube_user_pod_memory"`
	ClusterEnvironment      string `yaml:"clusterEnvironment" json:"cluster_environment"`

	ExecutorDockerImage string `yaml:"ExecutorDockerImage" json:"executor_docker_image,omitempty"`
	ExecutorDevBranch   string `yaml:"ExecutorDevBranch" json:"executor_development_branch,omitempty"`

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

func (*K8sConfig) Type() ManagerType {
	return K8sType
}

func (*ProcessConfig) Type() ManagerType {
	return ProcessType
}

func RegisterGobTypes() {
	gob.Register(&ProcessConfig{})
	gob.Register(&K8sConfig{})
	gob.Register(&WorkflowSpec{})
	gob.Register(&WorkflowRetentionSpec{})
}

func init() {
	RegisterGobTypes()
}
