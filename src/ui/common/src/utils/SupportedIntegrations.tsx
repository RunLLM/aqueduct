import React from 'react';

import {
  AirflowDialog,
  getAirflowValidationSchema,
} from '../components/integrations/dialogs/airflowDialog';
import {
  AthenaDialog,
  getAthenaValidationSchema,
} from '../components/integrations/dialogs/athenaDialog';
import {
  AWSDialog,
  getAWSValidationSchema,
} from '../components/integrations/dialogs/awsDialog';
import AzureDialog, {
  getAzureValidationSchema,
} from '../components/integrations/dialogs/azureDialog';
import {
  BigQueryDialog,
  getBigQueryValidationSchema,
} from '../components/integrations/dialogs/bigqueryDialog';
import { CondaDialog } from '../components/integrations/dialogs/condaDialog';
import {
  DatabricksDialog,
  getDatabricksValidationSchema,
} from '../components/integrations/dialogs/databricksDialog';
import {
  ECRDialog,
  getECRValidationSchema,
} from '../components/integrations/dialogs/ecrDialog';
import {
  EmailDialog,
  getEmailValidationSchema,
} from '../components/integrations/dialogs/emailDialog';
import GCPDialog, {
  getGCPValidationSchema,
} from '../components/integrations/dialogs/gcpDialog';
import {
  GCSDialog,
  getGCSValidationSchema,
} from '../components/integrations/dialogs/gcsDialog';
import {
  getLambdaValidationSchema,
  LambdaDialog,
} from '../components/integrations/dialogs/lambdaDialog';
import {
  getMariaDBValidationSchema,
  MariaDbDialog,
} from '../components/integrations/dialogs/mariadbDialog';
import {
  getMongoDBValidationSchema,
  MongoDBDialog,
} from '../components/integrations/dialogs/mongoDbDialog';
import {
  getMySQLValidationSchema,
  MysqlDialog,
} from '../components/integrations/dialogs/mysqlDialog';
import OnDemandKubernetesDialog, {
  getOnDemandKubernetesValidationSchema,
} from '../components/integrations/dialogs/onDemandKubernetesDialog';
import {
  getPostgresValidationSchema,
  PostgresDialog,
} from '../components/integrations/dialogs/postgresDialog';
import {
  getRedshiftValidationSchema,
  RedshiftDialog,
} from '../components/integrations/dialogs/redshiftDialog';
import {
  getS3ValidationSchema,
  S3Dialog,
} from '../components/integrations/dialogs/s3Dialog';
import {
  getSlackValidationSchema,
  SlackDialog,
} from '../components/integrations/dialogs/slackDialog';
import {
  getSnowflakeValidationSchema,
  SnowflakeDialog,
} from '../components/integrations/dialogs/snowflakeDialog';
import {
  getSparkValidationSchema,
  SparkDialog,
} from '../components/integrations/dialogs/sparkDialog';
import {
  getSQLiteValidationSchema,
  SQLiteDialog,
} from '../components/integrations/dialogs/sqliteDialog';
import { AqueductDocsLink } from './docs';
import ExecutionStatus, { AWSCredentialType, ExecState } from './shared';

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

export function resourceStatus(integration: Integration): ExecutionStatus {
  // If an exec_state doesn't exist, we currently assume that it is a legacy resource that has succeeded.
  let status = integration.exec_state?.status || ExecutionStatus.Succeeded;

  // For Aqueduct compute, we'll also need to look at the status of any registered Conda.
  if (integration.service == "Aqueduct" && integration.exec_state.status == ExecutionStatus.Succeeded){
    const aqConfig = integration.config as AqueductComputeConfig
    if (aqConfig.conda_config_serialized) {
      const serialized_conda_exec_state = JSON.parse(aqConfig.conda_config_serialized)["exec_state"]
      const conda_exec_state = JSON.parse(serialized_conda_exec_state) as ExecState
      status = conda_exec_state.status
    }
  }
  return status
}

export function isCondaRegistered(aqConfig: AqueductComputeConfig): boolean {
  return aqConfig.conda_config_serialized != undefined
}

export type Integration = {
  id: string;
  service: Service;
  name: string;
  config: IntegrationConfig;
  createdAt: number;
  exec_state: ExecState;
};

export type AqueductComputeConfig = {
  // Either the python version of the server or the registered conda field should be set,
  // but not both.
  python_version?: string;

  // Deserialize this to obtain the `CondaConfig`.
  conda_config_serialized?: string;
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
import {
  IntegrationCategories,
  ServiceInfoMap,
  ServiceLogos,
} from './integrations';

const addingIntegrationLink = `${AqueductDocsLink}/integrations/adding-an-integration`;

export const SupportedIntegrations: ServiceInfoMap = {
  ['Postgres']: {
    logo: ServiceLogos['Postgres'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: addingIntegrationLink,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <PostgresDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: getPostgresValidationSchema(),
  },
  ['Snowflake']: {
    logo: ServiceLogos['Snowflake'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: addingIntegrationLink,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <SnowflakeDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: getSnowflakeValidationSchema(),
  },
  ['Redshift']: {
    logo: ServiceLogos['Redshift'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: addingIntegrationLink,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <RedshiftDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: getRedshiftValidationSchema(),
  },
  ['BigQuery']: {
    logo: ServiceLogos['BigQuery'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: `${addingIntegrationLink}/connecting-to-google-bigquery`,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <BigQueryDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: getBigQueryValidationSchema(),
  },
  ['MySQL']: {
    logo: ServiceLogos['MySQL'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: addingIntegrationLink,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <MysqlDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: getMySQLValidationSchema(),
  },
  ['MariaDB']: {
    logo: ServiceLogos['MariaDB'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: addingIntegrationLink,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <MariaDbDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: getMariaDBValidationSchema(),
  },
  ['S3']: {
    logo: ServiceLogos['S3'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: `${addingIntegrationLink}/connecting-to-aws-s3`,
    dialog: ({
      user,
      editMode,
      onCloseDialog,
      loading,
      disabled,
      setMigrateStorage,
    }) => (
      <S3Dialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
        setMigrateStorage={setMigrateStorage}
      />
    ),
    validationSchema: getS3ValidationSchema(),
  },
  ['GCS']: {
    logo: ServiceLogos['GCS'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: `${addingIntegrationLink}/connecting-to-google-cloud-storage`,
    dialog: ({
      user,
      editMode,
      onCloseDialog,
      loading,
      disabled,
      setMigrateStorage,
    }) => (
      <GCSDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
        setMigrateStorage={setMigrateStorage}
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
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => <div />,
    validationSchema: null,
  },
  ['Filesystem']: {
    logo: ServiceLogos['Aqueduct'],
    activated: true,
    category: IntegrationCategories.ARTIFACT_STORAGE,
    docs: addingIntegrationLink,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => null,
    validationSchema: getSQLiteValidationSchema(),
  },
  ['SQLite']: {
    logo: ServiceLogos['SQLite'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: addingIntegrationLink,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <SQLiteDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: getSQLiteValidationSchema(),
  },
  ['Athena']: {
    logo: ServiceLogos['Athena'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: addingIntegrationLink,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <AthenaDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: getAthenaValidationSchema(),
  },
  ['Airflow']: {
    logo: ServiceLogos['Airflow'],
    activated: true,
    category: IntegrationCategories.COMPUTE,
    docs: addingIntegrationLink,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <AirflowDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: getAirflowValidationSchema(),
  },
  ['Kubernetes']: {
    logo: ServiceLogos['Kubernetes'],
    activated: true,
    category: IntegrationCategories.COMPUTE,
    docs: addingIntegrationLink,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <OnDemandKubernetesDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: getOnDemandKubernetesValidationSchema(),
  },
  // ['Kubernetes']: {
  //   logo: ServiceLogos['Kubernetes'],
  //   activated: true,
  //   category: IntegrationCategories.COMPUTE,
  //   docs: `${addingIntegrationLink}/connecting-to-k8s-cluster`,
  //   dialog: ({ editMode, onCloseDialog, loading, disabled }) => (
  //     <KubernetesDialog
  //       editMode={editMode}
  //       onCloseDialog={onCloseDialog}
  //       loading={loading}
  //       disabled={disabled}
  //     />
  //   ),
  //   validationSchema: getKubernetesValidationSchema(),
  // },
  ['Lambda']: {
    logo: ServiceLogos['Lambda'],
    activated: true,
    category: IntegrationCategories.COMPUTE,
    docs: `${addingIntegrationLink}/connecting-to-aws-lambda`,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <LambdaDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: getLambdaValidationSchema(),
  },
  ['MongoDB']: {
    logo: ServiceLogos['MongoDB'],
    activated: true,
    category: IntegrationCategories.DATA,
    docs: addingIntegrationLink,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <MongoDBDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: getMongoDBValidationSchema(),
  },
  ['Conda']: {
    logo: ServiceLogos['Conda'],
    activated: true,
    category: IntegrationCategories.COMPUTE,
    docs: `${addingIntegrationLink}/connecting-to-conda`,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <CondaDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: null,
  },
  ['Databricks']: {
    logo: ServiceLogos['Databricks'],
    activated: true,
    category: IntegrationCategories.COMPUTE,
    docs: `${addingIntegrationLink}/connecting-to-databricks`,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <DatabricksDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: getDatabricksValidationSchema(),
  },
  ['Email']: {
    logo: ServiceLogos['Email'],
    activated: true,
    category: IntegrationCategories.NOTIFICATION,
    docs: `${AqueductDocsLink}/notifications/connecting-to-email`,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <EmailDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: getEmailValidationSchema(),
  },
  ['Slack']: {
    logo: ServiceLogos['Slack'],
    activated: true,
    category: IntegrationCategories.NOTIFICATION,
    docs: `${AqueductDocsLink}/notifications/connecting-to-slack`,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <SlackDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: getSlackValidationSchema(),
  },
  ['Spark']: {
    logo: ServiceLogos['Spark'],
    activated: true,
    category: IntegrationCategories.COMPUTE,
    docs: addingIntegrationLink,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <SparkDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: getSparkValidationSchema(),
  },
  // Not sure the difference between this one and the Amazon one below.
  ['AWS']: {
    logo: ServiceLogos['Kubernetes'],
    activated: true,
    category: IntegrationCategories.CLOUD,
    docs: addingIntegrationLink,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <AWSDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: getAWSValidationSchema(),
  },
  ['Amazon']: {
    logo: ServiceLogos['AWS'],
    activated: true,
    category: IntegrationCategories.CLOUD,
    docs: addingIntegrationLink,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <AWSDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: getAWSValidationSchema(),
  },
  ['GCP']: {
    logo: ServiceLogos['GCP'],
    activated: false,
    category: IntegrationCategories.CLOUD,
    docs: addingIntegrationLink,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <GCPDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: getGCPValidationSchema(),
  },
  ['Azure']: {
    logo: ServiceLogos['Azure'],
    activated: false,
    category: IntegrationCategories.CLOUD,
    docs: addingIntegrationLink,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <AzureDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: getAzureValidationSchema(),
  },
  ['ECR']: {
    logo: ServiceLogos['ECR'],
    activated: true,
    category: IntegrationCategories.CONTAINER_REGISTRY,
    docs: addingIntegrationLink,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => (
      <ECRDialog
        user={user}
        editMode={editMode}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: getECRValidationSchema(),
  },
};

export default SupportedIntegrations;
