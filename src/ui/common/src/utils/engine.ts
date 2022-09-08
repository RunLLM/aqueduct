export enum EngineType {
  Aqueduct = 'aqueduct',
  Airflow = 'airflow',
  K8s = 'k8s',
}

export type AqueductConfig = Record<string, never>; // empty object

export type K8sConfig = {
  integration_id: string;
};

export type AirflowConfig = {
  integration_id: string;
  dag_id: string;
  operator_to_task: string;
  operator_metadata_path_prefix: string;
  artifact_content_path_prefix: string;
  artifact_metadata_path_prefix: string;
};

export type EngineConfig = {
  type: EngineType;
  aqueduct_config?: AqueductConfig;
  k8s_config?: K8sConfig;
  airflow_config?: AirflowConfig;
};
