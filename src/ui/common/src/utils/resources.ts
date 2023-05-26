import { ObjectSchema } from 'yup';

import { apiAddress } from '../components/hooks/useAqueductConsts';
import UserProfile from './auth';
import ExecutionStatus, { AWSCredentialType, ExecState } from './shared';

export const aqueductDemoName = 'Demo';
export const aqueductComputeName = 'Aqueduct Server';
export const aqueductStorageName = 'Filesystem';

export function isBuiltinResource(resource: Resource): boolean {
  return (
    resource.name === aqueductDemoName ||
    resource.name == aqueductComputeName ||
    resource.name == aqueductStorageName
  );
}

export function isNotificationResource(resource: Resource): boolean {
  return resource?.service == 'Email' || resource?.service == 'Slack';
}

export function resourceExecState(resource: Resource): ExecState {
  // For Aqueduct compute, we'll also need to look at the status of any registered Conda.
  if (
    resource.service == 'Aqueduct' &&
    isCondaRegistered(resource) &&
    resource.exec_state.status == ExecutionStatus.Succeeded
  ) {
    const aqConfig = resource.config as AqueductComputeConfig;
    if (aqConfig.conda_config_serialized) {
      const serialized_conda_exec_state = JSON.parse(
        aqConfig.conda_config_serialized
      )['exec_state'];
      const conda_exec_state = JSON.parse(
        serialized_conda_exec_state
      ) as ExecState;
      return conda_exec_state;
    }
  }

  return resource.exec_state || { status: ExecutionStatus.Succeeded };
}

// The only resource that does not necessarily display the same service type as
// on the resource itself is Conda.
export function resolveDisplayService(resource: Resource): Service {
  if (resource.service === 'Aqueduct') {
    const aqConfig = resource.config as AqueductComputeConfig;
    if (aqConfig.conda_config_serialized) {
      return 'Conda';
    }
  }
  return resource.service;
}

export function isCondaRegistered(resource: Resource): boolean {
  const aqConfig = resource.config as AqueductComputeConfig;
  return aqConfig?.conda_config_serialized != undefined;
}

export type Resource = {
  id: string;
  service: Service;
  name: string;
  config: ResourceConfig;
  createdAt: number;
  exec_state: ExecState;
};

export type AqueductComputeConfig = {
  // Either the python version of the server or the conda fields should be set,
  // but not both.
  python_version?: string;

  // Deserialize this to obtain the `CondaConfig`.
  conda_config_serialized?: string;
  conda_resource_id?: string;
  conda_resource_name?: string;
};

export type CondaConfig = {
  exec_state: string;
  conda_path: string;
};

export type PostgresConfig = {
  host: string;
  port: string;
  database: string;
  username: string;
  password?: string;
};

export type SnowflakeConfig = {
  account_identifier: string;
  warehouse: string;
  database: string;
  schema: string;
  username: string;
  password?: string;
  role: string;
};

export type RedshiftConfig = {
  host: string;
  port: string;
  database: string;
  username: string;
  password?: string;
};

export type BigQueryConfig = {
  project_id: string;
  service_account_credentials?: string;
};

export type MySqlConfig = {
  host: string;
  port: string;
  database: string;
  username: string;
  password?: string;
};

export type MariaDbConfig = {
  host: string;
  port: string;
  database: string;
  username: string;
  password?: string;
};

export type MongoDBConfig = {
  auth_uri: string;
  database: string;
};

export type SqlServerConfig = {
  host: string;
  port: string;
  database: string;
  username: string;
  password?: string;
};

export type GoogleSheetsConfig = {
  email: string;
  code?: string;
};

export type GithubConfig = {
  code?: string;
};

export type SalesforceConfig = {
  instance_url?: string;
  code?: string;
};

export enum DynamicEngineType {
  K8s = 'k8s',
}

export type S3Config = {
  type: AWSCredentialType;
  bucket: string;
  region: string;

  // If set, expected to be in the format `path/to/dir/`
  root_dir: string;
  access_key_id: string;
  secret_access_key: string;
  config_file_path: string;
  config_file_content: string;
  config_file_profile: string;
  use_as_storage: string;
};

export type AthenaConfig = {
  type: AWSCredentialType;
  access_key_id: string;
  secret_access_key: string;
  region: string;
  config_file_path: string;
  config_file_content: string;
  config_file_profile: string;
  database: string;
  output_location: string;
};

export type GCSConfig = {
  bucket: string;
  service_account_credentials?: string;
  use_as_storage: string;
};

export type AqueductDemoConfig = Record<string, never>;

export type AirflowConfig = {
  host: string;
  username: string;
  password: string;
  s3_credentials_path: string;
  s3_credentials_profile: string;
};

export type SQLiteConfig = {
  database: string;
};

export type KubernetesConfig = {
  kubeconfig_path: string;
  cluster_name: string;
  use_same_cluster: string;
};

export type LambdaConfig = {
  role_arn: string;
  exec_state: string;
};

export type DatabricksConfig = {
  workspace_url: string;
  access_token: string;
  s3_instance_profile_arn: string;
  instance_pool_id: string;
};

export type NotificationResourceConfig = {
  level: string;
  enabled: 'true' | 'false'; // this has to be string to fit backend requirements.
};

export type EmailConfig = {
  host: string;
  port: string;
  user: string;
  password: string;
  targets_serialized: string; // This should be a serialized list
} & NotificationResourceConfig;

export type SlackConfig = {
  token: string;
  channels_serialized: string;
} & NotificationResourceConfig;

export type SparkConfig = {
  livy_server_url: string;
};

export type AWSConfig = {
  type: AWSCredentialType;
  region: string;
  access_key_id: string;
  secret_access_key: string;
  config_file_path: string;
  config_file_profile: string;
  k8s_serialized: string;
};

export type DynamicK8sConfig = {
  keepalive: string;
  cpu_node_type: string;
  gpu_node_type: string;
  min_cpu_node: string;
  max_cpu_node: string;
  min_gpu_node: string;
  max_gpu_node: string;
};

export type ECRConfig = {
  type: AWSCredentialType;
  region: string;
  access_key_id: string;
  secret_access_key: string;
  config_file_path: string;
  config_file_profile: string;
};

export type FilesystemConfig = {
  location: string;
};

export type ResourceConfig =
  | PostgresConfig
  | SnowflakeConfig
  | RedshiftConfig
  | BigQueryConfig
  | MySqlConfig
  | MariaDbConfig
  | SqlServerConfig
  | GoogleSheetsConfig
  | SalesforceConfig
  | S3Config
  | AthenaConfig
  | GCSConfig
  | AqueductComputeConfig
  | AqueductDemoConfig
  | AirflowConfig
  | KubernetesConfig
  | LambdaConfig
  | CondaConfig
  | DatabricksConfig
  | EmailConfig
  | SlackConfig
  | SparkConfig
  | AWSConfig
  | MongoDBConfig
  | FilesystemConfig;

export type FullResourceConfig = { name: string } & ResourceConfig;

export type Service =
  | 'Aqueduct'
  | 'Postgres'
  | 'Snowflake'
  | 'Redshift'
  | 'BigQuery'
  | 'MySQL'
  | 'MariaDB'
  | 'S3'
  | 'Athena'
  | 'GCS'
  | 'Airflow'
  | 'Kubernetes'
  | 'SQLite'
  | 'Lambda'
  | 'Google Sheets'
  | 'MongoDB'
  | 'Conda'
  | 'Databricks'
  | 'Email'
  | 'Slack'
  | 'Spark'
  | 'AWS'
  | 'Amazon'
  | 'GCP'
  | 'Azure'
  | 'ECR'
  | 'GAR'
  | 'Filesystem';

export type Info = {
  logo: string;
  activated: boolean;
  category: string;
  docs: string;
  dialog?: React.FC<ResourceDialogProps<ResourceConfig>>;
  // TODO: figure out typescript type for yup schema
  // This may be useful: https://stackoverflow.com/questions/66171196/how-to-use-yups-object-shape-with-typescript
  validationSchema: (editMode: boolean) => ObjectSchema<any>;
};

export type ServiceInfoMap = {
  [key: string]: Info;
};

export type ServiceLogo = {
  [key: Service]: string;
};

export type FileData = {
  name: string;
  data: string;
};

export type CSVConfig = {
  name: string;
  csv: FileData;
};

export async function addTable(
  user: UserProfile,
  resourceId: string,
  config: CSVConfig
): Promise<void> {
  const res = await fetch(`${apiAddress}/api/resource/${resourceId}/create`, {
    method: 'POST',
    headers: {
      'api-key': user.apiKey,
      'table-name': config.name,
    },
    body: config.csv.data,
  });

  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.error);
  }
}

// S3 bucket folder for Aqueduct logos.
const logoBucket =
  'https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/logos';

// S3 bucket folder for Resource logos.
const resourceLogosBucket =
  'https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/logos/resources';

export const ResourceCategories = {
  DATA: 'data',
  COMPUTE: 'compute',
  CLOUD: 'cloud',
  CONTAINER_REGISTRY: 'container_registry',
  NOTIFICATION: 'notification',
  ARTIFACT_STORAGE: 'artifact_storage',
};

export const ServiceLogos: ServiceLogo = {
  ['Aqueduct']: `${logoBucket}/aqueduct-logo-two-tone/small/2x/aqueduct-logo-two-tone-small%402x.png`,
  ['Postgres']: `${resourceLogosBucket}/440px-Postgresql_elephant.svg.png`,
  ['Snowflake']: `${resourceLogosBucket}/51-513957_periscope-data-partners-snowflake-computing-logo.png`,
  ['Redshift']: `${resourceLogosBucket}/amazon-redshift.png`,
  ['BigQuery']: `${resourceLogosBucket}/google-bigquery-logo-1.svg`,
  ['MySQL']: `${resourceLogosBucket}/mysql.png`,
  ['MariaDB']: `${resourceLogosBucket}/mariadb.png`,
  ['S3']: `${resourceLogosBucket}/s3.png`,
  ['GCS']: `${resourceLogosBucket}/google-cloud-storage.png`,
  ['SQLite']: `${resourceLogosBucket}/sqlite-square-icon-256x256.png`,
  ['Athena']: `${resourceLogosBucket}/athena.png`,
  ['Airflow']: `${resourceLogosBucket}/airflow.png`,
  ['Kubernetes']: `${resourceLogosBucket}/kubernetes.png`,
  ['Lambda']: `${resourceLogosBucket}/lambda.png`,
  ['MongoDB']: `${resourceLogosBucket}/mongo.png`,
  ['Conda']: `${resourceLogosBucket}/conda.png`,
  ['Databricks']: `${resourceLogosBucket}/databricks_logo.png`,
  ['Email']: `${resourceLogosBucket}/email.png`,
  ['Slack']: `${resourceLogosBucket}/slack.png`,
  ['Spark']: `${resourceLogosBucket}/spark-logo-only.png`,
  ['AWS']: `${resourceLogosBucket}/aws-logo-trademark.png`,
  ['GCP']: `${resourceLogosBucket}/gcp.png`,
  ['Azure']: `${resourceLogosBucket}/azure.png`,

  // TODO(ENG-2301): Once task is addressed, remove this duplicate entry.
  ['K8s']: `${resourceLogosBucket}/kubernetes.png`,

  ['ECR']: `${resourceLogosBucket}/ecr.png`,
  ['GAR']: `${resourceLogosBucket}/gar.png`,
};

export type ResourceDialogProps<ResourceType> = {
  user: UserProfile;
  resourceToEdit?: ResourceType;
  onCloseDialog?: () => void;
  loading: boolean;
  disabled: boolean;
  setMigrateStorage?: (migrate: boolean) => void;
};

// Helper function to format resource service
export function formatService(service: string): string {
  service = service.toLowerCase();
  return service.replace(/ /g, '_');
}
