export enum EngineType {
  Aqueduct = 'aqueduct',
  Airflow = 'airflow',
  K8s = 'k8s',
  Lambda = 'lambda',
}

export type EngineConfig = {
  type: EngineType;
  aqueduct_config?: AqueductConfig;
  airflow_config?: AirflowConfig;
  k8s_config?: K8sConfig;
  lambda_config?: LambdaConfig;
};

export type AqueductConfig = Record<string, never>;

export type AirflowConfig = {
  integration_id: string;
  matches_airflow: boolean;
};

export type K8sConfig = {
  integration_id: string;
};

export type LambdaConfig = {
  integration_id: string;
};
