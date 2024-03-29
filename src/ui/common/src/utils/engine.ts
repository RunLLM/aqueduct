import { Service } from './resources';

export enum EngineType {
  AqueductConda = 'aqueduct_conda',
  Aqueduct = 'aqueduct',
  Airflow = 'airflow',
  Databricks = 'databricks',
  Spark = 'spark',
  K8s = 'k8s',
  Lambda = 'lambda',
}

export const EngineTypeToService: { [engineType: string]: Service } = {
  [EngineType.Aqueduct]: 'Aqueduct',
  [EngineType.AqueductConda]: 'Conda',
  [EngineType.Airflow]: 'Airflow',
  [EngineType.Databricks]: 'Databricks',
  [EngineType.Spark]: 'Spark',
  [EngineType.K8s]: 'Kubernetes',
  [EngineType.Lambda]: 'Lambda',
};

export type EngineConfig = {
  type: EngineType;
  aqueduct_conda_config?: AqueductCondaConfig;
  aqueduct_config?: AqueductConfig;
  airflow_config?: AirflowConfig;
  k8s_config?: K8sConfig;
  lambda_config?: LambdaConfig;
  databricks_config?: DatabricksConfig;
  spark_config?: SparkConfig;
};

export type EngineWithResource = {
  resource_id: string;
};

export type AqueductConfig = Record<string, never>;

export type AqueductCondaConfig = {
  env: string;
};

export type AirflowConfig = EngineWithResource & {
  matches_airflow: boolean;
};

export type DatabricksConfig = EngineWithResource;

export type SparkConfig = EngineWithResource & {
  environment_path_uri: string;
};

export type K8sConfig = EngineWithResource;

export type LambdaConfig = EngineWithResource;
