export enum EngineType {
  Aqueduct = 'aqueduct',
  Airflow = 'airflow',
  K8s = 'k8s',
  Lambda = 'lambda',
}

export type EngineConfig = {
  type: EngineType;
  airflow_config?: AirflowConfig;
};

export type AirflowConfig = {
  matches_airflow: boolean;
};


