import { apiAddress } from '../components/hooks/useAqueductConsts';
import UserProfile from './auth';

export const aqueductDemoName = 'aqueduct_demo';

export function isDemo(integration: Integration): boolean {
  return integration.name === aqueductDemoName;
}

export type Integration = {
  id: string;
  service: Service;
  name: string;
  config: IntegrationConfig;
  createdAt: number;
  validated: boolean;
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
  username: string;
  password?: string;
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

export enum AWSCredentialType {
  AccessKey = 'access_key',
  ConfigFilePath = 'config_file_path',
  ConfigFileContent = 'config_file_content',
}

export type S3Config = {
  type: AWSCredentialType;
  bucket: string;
  region: string;
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
};

export type DatabricksConfig = {
  workspace_url: string;
  access_token: string;
  s3_instance_profile_arn: string;
};

export type EmailConfig = {
  host: string;
  port: string;
  user: string;
  password: string;
  targets_serialized: string; // This should be a serialized list
  level: string;
};

export type SlackConfig = {
  token: string;
  channels_serialized: string;
  level: string;
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
  | SlackConfig;

export type Service =
  | 'Postgres'
  | 'Snowflake'
  | 'Redshift'
  | 'BigQuery'
  | 'MySQL'
  | 'MariaDB'
  | 'S3'
  | 'Athena'
  | 'CSV'
  | 'GCS'
  | 'Aqueduct Demo'
  | 'Airflow'
  | 'Kubernetes'
  | 'SQLite'
  | 'Lambda'
  | 'Google Sheets'
  | 'MongoDB'
  | 'Conda'
  | 'Databricks'
  | 'Email'
  | 'Slack';

export type Info = {
  logo: string;
  activated: boolean;
  category: string;
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
  NOTIFICATION: 'notification',
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
  ['Aqueduct Demo']: `/assets/aqueduct.png`,
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
};

export const SupportedIntegrations: ServiceInfoMap = {
  ['Postgres']: {
    logo: ServiceLogos['Postgres'],
    activated: true,
    category: IntegrationCategories.DATA,
  },
  ['Snowflake']: {
    logo: ServiceLogos['Snowflake'],
    activated: true,
    category: IntegrationCategories.DATA,
  },
  ['Redshift']: {
    logo: ServiceLogos['Redshift'],
    activated: true,
    category: IntegrationCategories.DATA,
  },
  ['BigQuery']: {
    logo: ServiceLogos['BigQuery'],
    activated: true,
    category: IntegrationCategories.DATA,
  },
  ['MySQL']: {
    logo: ServiceLogos['MySQL'],
    activated: true,
    category: IntegrationCategories.DATA,
  },
  ['MariaDB']: {
    logo: ServiceLogos['MariaDB'],
    activated: true,
    category: IntegrationCategories.DATA,
  },
  ['S3']: {
    logo: ServiceLogos['S3'],
    activated: true,
    category: IntegrationCategories.DATA,
  },
  ['GCS']: {
    logo: ServiceLogos['GCS'],
    activated: true,
    category: IntegrationCategories.DATA,
  },
  ['Aqueduct Demo']: {
    logo: ServiceLogos['Aqueduct Demo'],
    activated: true,
    category: IntegrationCategories.DATA,
  },
  ['SQLite']: {
    logo: ServiceLogos['SQLite'],
    activated: true,
    category: IntegrationCategories.DATA,
  },
  ['Athena']: {
    logo: ServiceLogos['Athena'],
    activated: true,
    category: IntegrationCategories.DATA,
  },
  ['Airflow']: {
    logo: ServiceLogos['Airflow'],
    activated: true,
    category: IntegrationCategories.COMPUTE,
  },
  ['Kubernetes']: {
    logo: ServiceLogos['Kubernetes'],
    activated: true,
    category: IntegrationCategories.COMPUTE,
  },
  ['Lambda']: {
    logo: ServiceLogos['Lambda'],
    activated: true,
    category: IntegrationCategories.COMPUTE,
  },
  ['MongoDB']: {
    logo: ServiceLogos['MongoDB'],
    activated: true,
    category: IntegrationCategories.DATA,
  },
  ['Conda']: {
    logo: ServiceLogos['Conda'],
    activated: true,
    category: IntegrationCategories.COMPUTE,
  },
  ['Databricks']: {
    logo: ServiceLogos['Databricks'],
    activated: false,
    category: IntegrationCategories.COMPUTE,
  },
  ['Email']: {
    logo: ServiceLogos['Email'],
    activated: true,
    category: IntegrationCategories.NOTIFICATION,
  },
  ['Slack']: {
    logo: ServiceLogos['Slack'],
    activated: true,
    category: IntegrationCategories.NOTIFICATION,
  },
};

// Helper function to format integration service
export function formatService(service: string): string {
  service = service.toLowerCase();
  return service.replace(/ /g, '_');
}
