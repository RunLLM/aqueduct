import { useAqueductConsts } from '../components/hooks/useAqueductConsts';
import UserProfile from './auth';

const { apiAddress } = useAqueductConsts();

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
  | KubernetesConfig;

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
  | 'SQLite';

type Info = {
  logo: string;
  activated: boolean;
  category: string;
};

export type ServiceInfoMap = {
  [key: string]: Info;
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

export async function fetchRepos(
  user: UserProfile
): Promise<[string[], string]> {
  try {
    const res = await fetch(`${apiAddress}/api/integrations/github/repos`, {
      method: 'GET',
      headers: {
        'api-key': user.apiKey,
      },
    });

    if (!res.ok) {
      return [[], await res.text()];
    }

    const body = await res.json();
    return [body.repos, ''];
  } catch (err) {
    return [[], err];
  }
}

export async function fetchBranches(
  user: UserProfile,
  repo: string
): Promise<[string[], string]> {
  try {
    const res = await fetch(`${apiAddress}/api/integrations/github/branches`, {
      method: 'GET',
      headers: {
        'api-key': user.apiKey,
        'github-repo': repo,
      },
    });

    if (!res.ok) {
      return [[], await res.text()];
    }

    const body = await res.json();
    return [body.branches, ''];
  } catch (err) {
    return [[], err];
  }
}

export const SupportedIntegrations: ServiceInfoMap = {
  ['Postgres']: {
    logo: 'https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/pages/integrations/440px-Postgresql_elephant.svg.png',
    activated: true,
    category: 'data',
  },
  ['Snowflake']: {
    logo: 'https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/pages/integrations/51-513957_periscope-data-partners-snowflake-computing-logo.png',
    activated: true,
    category: 'data',
  },
  ['Redshift']: {
    logo: 'https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/pages/integrations/amazon-redshift.png',
    activated: true,
    category: 'data',
  },
  ['BigQuery']: {
    logo: 'https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/pages/integrations/google-bigquery-logo-1.svg',
    activated: true,
    category: 'data',
  },
  ['MySQL']: {
    logo: 'https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/pages/integrations/mysql.png',
    activated: true,
    category: 'data',
  },
  ['MariaDB']: {
    logo: 'https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/pages/integrations/mariadb.png',
    activated: true,
    category: 'data',
  },
  ['S3']: {
    logo: 'https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/pages/integrations/s3.png',
    activated: true,
    category: 'data',
  },
  ['GCS']: {
    logo: 'https://spiral-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/pages/integrations/google-cloud-storage.png',
    activated: true,
    category: 'data',
  },
  ['Aqueduct Demo']: {
    logo: '/assets/aqueduct.png',
    activated: true,
    category: 'data',
  },
  ['SQLite']: {
    logo: 'https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/pages/integrations/sqlite-square-icon-256x256.png',
    activated: true,
    category: 'data',
  },
  ['Athena']: {
    logo: 'https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/pages/integrations/athena.png',
    activated: true,
    category: 'data',
  },
  ['Airflow']: {
    logo: 'https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/pages/integrations/airflow.png',
    activated: false,
    category: 'compute',
  },
  ['Kubernetes']: {
    logo: 'https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/pages/integrations/kubernetes.png',
    activated: true,
    category: 'compute',
  },
};

// Helper function to format integration service
export function formatService(service: string): string {
  service = service.toLowerCase();
  return service.replace(/ /g, '_');
}
