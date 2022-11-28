package shared

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/google/uuid"
)

type EngineType string

const (
	AqueductEngineType EngineType = "aqueduct"
	AirflowEngineType  EngineType = "airflow"
	K8sEngineType      EngineType = "k8s"
	LambdaEngineType   EngineType = "lambda"
)

type EngineConfig struct {
	Type           EngineType      `yaml:"type" json:"type"`
	AqueductConfig *AqueductConfig `yaml:"aqueductConfig" json:"aqueduct_config,omitempty"`
	AirflowConfig  *AirflowConfig  `yaml:"airflowConfig" json:"airflow_config,omitempty"`
	K8sConfig      *K8sConfig      `yaml:"k8sConfig" json:"k8s_config,omitempty"`
	LambdaConfig   *LambdaConfig   `yaml:"lambdaConfig" json:"lambda_config,omitempty"`
}

type AqueductConfig struct{}

type AirflowConfig struct {
	IntegrationID uuid.UUID `json:"integration_id"  yaml:"integration_id"`
	DagID         string    `json:"dag_id"  yaml:"dag_id"`
	// MatchesAirflow indicates whether this DAG matches the current DAG registered with Airflow
	MatchesAirflow             bool                 `json:"matches_airflow"  yaml:"matches_airflow"`
	OperatorToTask             map[uuid.UUID]string `json:"operator_to_task"  yaml:"operator_to_task"`
	OperatorMetadataPathPrefix map[uuid.UUID]string `json:"operator_metadata_path_prefix"  yaml:"operator_metadata_path_prefix"`
	ArtifactContentPathPrefix  map[uuid.UUID]string `json:"artifact_content_path_prefix"  yaml:"artifact_content_path_prefix"`
	ArtifactMetadataPathPrefix map[uuid.UUID]string `json:"artifact_metadata_path_prefix"  yaml:"artifact_metadata_path_prefix"`
}

type K8sConfig struct {
	IntegrationID uuid.UUID `json:"integration_id"  yaml:"integration_id"`
}

type LambdaConfig struct {
	IntegrationID uuid.UUID `json:"integration_id"  yaml:"integration_id"`
}

func (e *EngineConfig) Scan(value interface{}) error {
	return utils.ScanJSONB(value, e)
}

func (e *EngineConfig) Value() (driver.Value, error) {
	return utils.ValueJSONB(*e)
}
