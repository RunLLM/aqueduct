import { apiAddress } from '../components/hooks/useAqueductConsts';
import UserProfile from './auth';
import { AWSCredentialType, ExecState } from './shared';

export const aqueductDemoName = 'Demo';
export const aqueductComputeName = 'Aqueduct Server';
export const aqueductStorageName = 'Filesystem';

export function isBuiltinIntegration(integration: Integration): boolean {
  return (
    integration.name === aqueductDemoName ||
    integration.name == aqueductComputeName ||
    integration.name == aqueductStorageName
  );
}

export function isNotificationIntegration(integration: Integration): boolean {
  return integration.service == 'Email' || integration.service == 'Slack';
}

// Certain integrations have no configuration fields to show.
export function hasConfigFieldsToShow(integration: Integration): boolean {
  return (
    integration.service !== 'Conda' && integration.name !== aqueductComputeName
  );
}

export type Integration = {
  id: string;
  service: Service;
  name: string;
  config: IntegrationConfig;
  createdAt: number;
  exec_state: ExecState;
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

export type NotificationIntegrationConfig = {
  level: string;
  enabled: 'true' | 'false'; // this has to be string to fit backend requirements.
};

export type EmailConfig = {
  host: string;
  port: string;
  user: string;
  password: string;
  targets_serialized: string; // This should be a serialized list
} & NotificationIntegrationConfig;

export type SlackConfig = {
  token: string;
  channels_serialized: string;
} & NotificationIntegrationConfig;

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

export type IntegrationConfig =
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
  | 'Filesystem';

export type Info = {
  logo: string;
  activated: boolean;
  category: string;
  docs: string;
  dialog?: React.FC<IntegrationDialogProps>;
  // TODO: figure out typescript type for yup schema
  // This may be useful: https://stackoverflow.com/questions/66171196/how-to-use-yups-object-shape-with-typescript
  validationSchema: any;
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
  integrationId: string,
  config: CSVConfig
): Promise<void> {
  const res = await fetch(
    `${apiAddress}/api/integration/${integrationId}/create`,
    {
      method: 'POST',
      headers: {
        'api-key': user.apiKey,
        'table-name': config.name,
      },
      body: config.csv.data,
    }
  );

  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.error);
  }
}

// S3 bucket folder for Aqueduct logos.
const logoBucket =
  'https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/logos';

// S3 bucket folder for Integration logos.
const integrationLogosBucket =
  'https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/pages/integrations';

export const IntegrationCategories = {
  DATA: 'data',
  COMPUTE: 'compute',
  CLOUD: 'cloud',
  CONTAINER_REGISTRY: 'container_registry',
  NOTIFICATION: 'notification',
  ARTIFACT_STORAGE: 'artifact_storage',
};

export const ServiceLogos: ServiceLogo = {
  ['Aqueduct']: `${logoBucket}/aqueduct-logo-two-tone/small/2x/aqueduct-logo-two-tone-small%402x.png`,
  ['Postgres']: `${integrationLogosBucket}/440px-Postgresql_elephant.svg.png`,
  ['Snowflake']: `${integrationLogosBucket}/51-513957_periscope-data-partners-snowflake-computing-logo.png`,
  ['Redshift']: `${integrationLogosBucket}/amazon-redshift.png`,
  ['BigQuery']: `${integrationLogosBucket}/google-bigquery-logo-1.svg`,
  ['MySQL']: `${integrationLogosBucket}/mysql.png`,
  ['MariaDB']: `${integrationLogosBucket}/mariadb.png`,
  ['S3']: `${integrationLogosBucket}/s3.png`,
  ['GCS']: `${integrationLogosBucket}/google-cloud-storage.png`,
  ['SQLite']: `${integrationLogosBucket}/sqlite-square-icon-256x256.png`,
  ['Athena']: `${integrationLogosBucket}/athena.png`,
  ['Airflow']: `${integrationLogosBucket}/airflow.png`,
  ['Kubernetes']: `${integrationLogosBucket}/kubernetes.png`,
  ['Lambda']: `${integrationLogosBucket}/Lambda.png`,
  ['MongoDB']: `${integrationLogosBucket}/mongo.png`,
  ['Conda']: `${integrationLogosBucket}/conda.png`,
  ['Databricks']: `${integrationLogosBucket}/databricks_logo.png`,
  ['Email']: `${integrationLogosBucket}/email.png`,
  ['Slack']: `${integrationLogosBucket}/slack.png`,
  ['Spark']: `${integrationLogosBucket}/spark-logo-trademark.png`,
  ['AWS']: `${integrationLogosBucket}/aws-logo-trademark.png`,
  ['GCP']: `${integrationLogosBucket}/gcp.png`,
  ['Azure']: `${integrationLogosBucket}/azure.png`,

  // TODO(ENG-2301): Once task is addressed, remove this duplicate entry.
  ['K8s']: `${integrationLogosBucket}/kubernetes.png`,

  ['ECR']: `${integrationLogosBucket}/ecr.png`,
};

export type IntegrationDialogProps = {
  user: UserProfile;
  editMode?: boolean;
  onCloseDialog?: () => void;
  loading: boolean;
  disabled: boolean;
  setMigrateStorage?: (migrate: boolean) => void;
};

// Helper function to format integration service
export function formatService(service: string): string {
  service = service.toLowerCase();
  return service.replace(/ /g, '_');
}