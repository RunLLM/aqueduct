import { Service } from './integrations';

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

export type EngineWithIntegration = {
  integration_id: string;
};

export type AqueductConfig = Record<string, never>;

export type AqueductCondaConfig = {
  env: string;
};

export type AirflowConfig = EngineWithIntegration & {
  matches_airflow: boolean;
};

export type DatabricksConfig = EngineWithIntegration;

export type SparkConfig = EngineWithIntegration & {
  environment_path_uri: string;
};

export type K8sConfig = EngineWithIntegration;

export type LambdaConfig = EngineWithIntegration;
