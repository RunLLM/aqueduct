package shared

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/google/uuid"
)

type EngineType string

const (
	AqueductEngineType      EngineType = "aqueduct"
	AqueductCondaEngineType EngineType = "aqueduct_conda"
	AirflowEngineType       EngineType = "airflow"
	K8sEngineType           EngineType = "k8s"
	LambdaEngineType        EngineType = "lambda"
	DatabricksEngineType    EngineType = "databricks"
	SparkEngineType         EngineType = "spark"
)

type EngineConfig struct {
	Type                EngineType           `yaml:"type" json:"type"`
	AqueductConfig      *AqueductConfig      `yaml:"aqueductConfig" json:"aqueduct_config,omitempty"`
	AqueductCondaConfig *AqueductCondaConfig `yaml:"aqueductCondaConfig" json:"aqueduct_conda_config,omitempty"`
	AirflowConfig       *AirflowConfig       `yaml:"airflowConfig" json:"airflow_config,omitempty"`
	K8sConfig           *K8sConfig           `yaml:"k8sConfig" json:"k8s_config,omitempty"`
	LambdaConfig        *LambdaConfig        `yaml:"lambdaConfig" json:"lambda_config,omitempty"`
	DatabricksConfig    *DatabricksConfig    `yaml:"databricksConfig" json:"databricks_config,omitempty"`
	SparkConfig         *SparkConfig         `yaml:"sparkConfig" json:"spark_config,omitempty"`
}

type AqueductConfig struct{}

type AqueductCondaConfig struct {
	Env string `yaml:"env" json:"env"`
}

type AirflowConfig struct {
	ResourceID uuid.UUID `json:"integration_id"  yaml:"integration_id"`
	DagID      string    `json:"dag_id"  yaml:"dag_id"`
	// MatchesAirflow indicates whether this DAG matches the current DAG registered with Airflow
	MatchesAirflow             bool                 `json:"matches_airflow"  yaml:"matches_airflow"`
	OperatorToTask             map[uuid.UUID]string `json:"operator_to_task"  yaml:"operator_to_task"`
	OperatorMetadataPathPrefix map[uuid.UUID]string `json:"operator_metadata_path_prefix"  yaml:"operator_metadata_path_prefix"`
	ArtifactContentPathPrefix  map[uuid.UUID]string `json:"artifact_content_path_prefix"  yaml:"artifact_content_path_prefix"`
	ArtifactMetadataPathPrefix map[uuid.UUID]string `json:"artifact_metadata_path_prefix"  yaml:"artifact_metadata_path_prefix"`
}

type K8sConfig struct {
	ResourceID uuid.UUID `json:"integration_id"  yaml:"integration_id"`
}

type LambdaConfig struct {
	ResourceID uuid.UUID `json:"integration_id"  yaml:"integration_id"`
}

type DatabricksConfig struct {
	ResourceID uuid.UUID `json:"integration_id"  yaml:"integration_id"`
}

type SparkConfig struct {
	ResourceID uuid.UUID `json:"integration_id"  yaml:"integration_id"`
	// URI to the packaged environment. This is passed when creating and uploading the
	// environment during execution.
	EnvironmentPathURI string `yaml:"environmentPathUri" json:"environment_path_uri"`
}

func (e *EngineConfig) Scan(value interface{}) error {
	return utils.ScanJSONB(value, e)
}

func (e *EngineConfig) Value() (driver.Value, error) {
	return utils.ValueJSONB(*e)
}
