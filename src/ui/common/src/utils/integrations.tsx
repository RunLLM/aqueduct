import React from 'react';
import * as Yup from 'yup';

import {
  AWSDialog,
  BigQueryDialog,
  CondaDialog,
  DatabricksDialog,
  EmailDialog,
  MariaDbDialog,
  MongoDBDialog,
  MysqlDialog,
  PostgresDialog,
  RedshiftDialog,
  S3Dialog,
  SlackDialog,
  SnowflakeDialog,
  SparkDialog,
} from '..';
import { apiAddress } from '../components/hooks/useAqueductConsts';
import {
  AirflowDialog,
  getAirflowValidationSchema,
} from '../components/integrations/dialogs/airflowDialog';
import {
  AthenaDialog,
  getAthenaValidationSchema,
} from '../components/integrations/dialogs/athenaDialog';
import { getAWSValidationSchema } from '../components/integrations/dialogs/awsDialog';
import { getBigQueryValidationSchema } from '../components/integrations/dialogs/bigqueryDialog';
import { getDatabricksValidationSchema } from '../components/integrations/dialogs/databricksDialog';
import { getEmailValidationSchema } from '../components/integrations/dialogs/emailDialog';
import {
  GCSDialog,
  getGCSValidationSchema,
} from '../components/integrations/dialogs/gcsDialog';
import {
  getKubernetesValidationSchema,
  KubernetesDialog,
} from '../components/integrations/dialogs/kubernetesDialog';
import {
  getLambdaValidationSchema,
  LambdaDialog,
} from '../components/integrations/dialogs/lambdaDialog';
import { getMariaDBValidationSchema } from '../components/integrations/dialogs/mariadbDialog';
import { getMongoDBValidationSchema } from '../components/integrations/dialogs/mongoDbDialog';
import { getMySQLValidationSchema } from '../components/integrations/dialogs/mysqlDialog';
import { getPostgresValidationSchema } from '../components/integrations/dialogs/postgresDialog';
import { getRedshiftValidationSchema } from '../components/integrations/dialogs/redshiftDialog';
import { getS3ValidationSchema } from '../components/integrations/dialogs/s3Dialog';
import { getSlackValidationSchema } from '../components/integrations/dialogs/slackDialog';
import { getSnowflakeValidationSchema } from '../components/integrations/dialogs/snowflakeDialog';
import { getSparkValidationSchema } from '../components/integrations/dialogs/sparkDialog';
import {
  getSQLiteValidationSchema,
  SQLiteDialog,
} from '../components/integrations/dialogs/sqliteDialog';
import UserProfile from './auth';
import { AqueductDocsLink } from './docs';
import { ExecState } from './shared';

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

export enum AWSCredentialType {
  AccessKey = 'access_key',
  ConfigFilePath = 'config_file_path',
  ConfigFileContent = 'config_file_content',
}

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
  dialog: React.FC<IntegrationDialogProps>;
  // TODO: figure out typescript type for yup schema
  // This may be useful: https://stackoverflow.com/questions/66171196/how-to-use-yups-object-shape-with-typescript
  validationSchema: Yup.ObjectSchema<any>;
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

const addingIntegrationLink = `${AqueductDocsLink}/integrations/adding-an-integration`;

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
  //onUpdateField: (field: keyof AWSConfig, value: string) => void;
  //value?: AWSConfig;
  editMode?: boolean;
};

export const SupportedIntegrations: ServiceInfoMap = {
  ['Postgres']: {
    logo: ServiceLogos['Postgres'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: addingIntegrationLink,
    // What is the type of dialog?
    dialog: () => <PostgresDialog />,
    validationSchema: getPostgresValidationSchema(),
  },
  ['Snowflake']: {
    logo: ServiceLogos['Snowflake'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: addingIntegrationLink,
    dialog: () => <SnowflakeDialog />,
    validationSchema: getSnowflakeValidationSchema(),
  },
  ['Redshift']: {
    logo: ServiceLogos['Redshift'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: addingIntegrationLink,
    dialog: () => <RedshiftDialog />,
    validationSchema: getRedshiftValidationSchema(),
  },
  ['BigQuery']: {
    logo: ServiceLogos['BigQuery'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: `${addingIntegrationLink}/connecting-to-google-bigquery`,
    dialog: () => <BigQueryDialog />,
    validationSchema: getBigQueryValidationSchema(),
  },
  ['MySQL']: {
    logo: ServiceLogos['MySQL'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: addingIntegrationLink,
    dialog: () => <MysqlDialog />,
    validationSchema: getMySQLValidationSchema(),
  },
  ['MariaDB']: {
    logo: ServiceLogos['MariaDB'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: addingIntegrationLink,
    dialog: () => <MariaDbDialog />,
    validationSchema: getMariaDBValidationSchema(),
  },
  ['S3']: {
    logo: ServiceLogos['S3'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: `${addingIntegrationLink}/connecting-to-aws-s3`,
    // TODO: Figure out how to pass in setMigrateStorage to handle storage migrations here.
    // Believe that setMigrateStorage can just be a redux action called from within this dialog.
    dialog: () => (
      <S3Dialog
        setMigrateStorage={(param) => console.log('setMigrateStorage: ', param)}
      />
    ),
    validationSchema: getS3ValidationSchema(),
  },
  ['GCS']: {
    logo: ServiceLogos['GCS'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: `${addingIntegrationLink}/connecting-to-google-cloud-storage`,
    // TODO: Figure out SetMigrateStorage here. This can just be a redux action afaik.
    // Believe that setMigrateStorage can just be a redux action called from within this dialog.
    dialog: () => (
      <GCSDialog
        setMigrateStorage={(param) => console.log('setMigrateStorage: ', param)}
      />
    ),
    validationSchema: getGCSValidationSchema(),
  },
  ['Aqueduct']: {
    logo: ServiceLogos['Aqueduct'],
    activated: true,
    category: IntegrationCategories.COMPUTE,
    docs: addingIntegrationLink,
    // TODO: Figure out what to show here.
    dialog: () => <div />,
    validationSchema: null,
  },
  ['Filesystem']: {
    logo: ServiceLogos['Aqueduct'],
    activated: true,
    category: IntegrationCategories.ARTIFACT_STORAGE,
    docs: addingIntegrationLink,
  },
  ['SQLite']: {
    logo: ServiceLogos['SQLite'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: addingIntegrationLink,
    dialog: () => <SQLiteDialog />,
    validationSchema: getSQLiteValidationSchema(),
  },
  ['Athena']: {
    logo: ServiceLogos['Athena'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: addingIntegrationLink,
    dialog: () => <AthenaDialog />,
    validationSchema: getAthenaValidationSchema(),
  },
  ['Airflow']: {
    logo: ServiceLogos['Airflow'],
    activated: true,
    category: IntegrationCategories.COMPUTE,
    docs: addingIntegrationLink,
    dialog: () => <AirflowDialog />,
    validationSchema: getAirflowValidationSchema(),
  },
  ['Kubernetes']: {
    logo: ServiceLogos['Kubernetes'],
    activated: true,
    category: IntegrationCategories.COMPUTE,
    docs: `${addingIntegrationLink}/connecting-to-k8s-cluster`,
    dialog: () => <KubernetesDialog />,
    validationSchema: getKubernetesValidationSchema(),
  },
  ['Lambda']: {
    logo: ServiceLogos['Lambda'],
    activated: true,
    category: IntegrationCategories.COMPUTE,
    docs: `${addingIntegrationLink}/connecting-to-aws-lambda`,
    dialog: () => <LambdaDialog />,
    validationSchema: getLambdaValidationSchema(),
  },
  ['MongoDB']: {
    logo: ServiceLogos['MongoDB'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: addingIntegrationLink,
    dialog: () => <MongoDBDialog />,
    validationSchema: getMongoDBValidationSchema(),
  },
  ['Conda']: {
    logo: ServiceLogos['Conda'],
    activated: true,
    category: IntegrationCategories.COMPUTE,
    docs: `${addingIntegrationLink}/connecting-to-conda`,
    dialog: () => <CondaDialog />,
    validationSchema: null,
  },
  ['Databricks']: {
    logo: ServiceLogos['Databricks'],
    activated: true,
    category: IntegrationCategories.COMPUTE,
    docs: `${addingIntegrationLink}/connecting-to-databricks`,
    dialog: () => <DatabricksDialog />,
    validationSchema: getDatabricksValidationSchema(),
  },
  ['Email']: {
    logo: ServiceLogos['Email'],
    activated: true,
    category: IntegrationCategories.NOTIFICATION,
    docs: `${AqueductDocsLink}/notifications/connecting-to-email`,
    dialog: () => <EmailDialog />,
    validationSchema: getEmailValidationSchema(),
  },
  ['Slack']: {
    logo: ServiceLogos['Slack'],
    activated: true,
    category: IntegrationCategories.NOTIFICATION,
    docs: `${AqueductDocsLink}/notifications/connecting-to-slack`,
    dialog: () => <SlackDialog />,
    validationSchema: getSlackValidationSchema(),
  },
  ['Spark']: {
    logo: ServiceLogos['Spark'],
    activated: true,
    category: IntegrationCategories.COMPUTE,
    docs: addingIntegrationLink,
    dialog: () => <SparkDialog />,
    validationSchema: getSparkValidationSchema(),
  },
  ['AWS']: {
    logo: ServiceLogos['Kubernetes'],
    activated: true,
    category: IntegrationCategories.CLOUD,
    docs: addingIntegrationLink,
  },
  ['Amazon']: {
    logo: ServiceLogos['AWS'],
    activated: true,
    category: IntegrationCategories.CLOUD,
    docs: addingIntegrationLink,
    dialog: () => <AWSDialog />,
    validationSchema: getAWSValidationSchema(),
  },
  ['GCP']: {
    logo: ServiceLogos['GCP'],
    activated: false,
    category: IntegrationCategories.CLOUD,
    docs: addingIntegrationLink,
  },
  ['Azure']: {
    logo: ServiceLogos['Azure'],
    activated: false,
    category: IntegrationCategories.CLOUD,
    docs: addingIntegrationLink,
  },
  ['ECR']: {
    logo: ServiceLogos['ECR'],
    activated: true,
    category: IntegrationCategories.CONTAINER_REGISTRY,
    docs: addingIntegrationLink,
  },
};

// Helper function to format integration service
export function formatService(service: string): string {
  service = service.toLowerCase();
  return service.replace(/ /g, '_');
}
