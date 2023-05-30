import React from 'react';

import {
  AirflowDialog,
  getAirflowValidationSchema,
} from '../components/resources/dialogs/airflowDialog';
import {
  AthenaDialog,
  getAthenaValidationSchema,
} from '../components/resources/dialogs/athenaDialog';
import {
  AWSDialog,
  getAWSValidationSchema,
} from '../components/resources/dialogs/awsDialog';
import AzureDialog, {
  getAzureValidationSchema,
} from '../components/resources/dialogs/azureDialog';
import {
  BigQueryDialog,
  getBigQueryValidationSchema,
} from '../components/resources/dialogs/bigqueryDialog';
import {
  CondaDialog,
  getCondaValidationSchema,
} from '../components/resources/dialogs/condaDialog';
import {
  DatabricksDialog,
  getDatabricksValidationSchema,
} from '../components/resources/dialogs/databricksDialog';
import {
  ECRDialog,
  getECRValidationSchema,
} from '../components/resources/dialogs/ecrDialog';
import {
  EmailDialog,
  getEmailValidationSchema,
} from '../components/resources/dialogs/emailDialog';
import {
  GARDialog,
  getGARValidationSchema,
} from '../components/resources/dialogs/garDialog';
import GCPDialog, {
  getGCPValidationSchema,
} from '../components/resources/dialogs/gcpDialog';
import {
  GCSDialog,
  getGCSValidationSchema,
} from '../components/resources/dialogs/gcsDialog';
import {
  getLambdaValidationSchema,
  LambdaDialog,
} from '../components/resources/dialogs/lambdaDialog';
import {
  getMariaDBValidationSchema,
  MariaDbDialog,
} from '../components/resources/dialogs/mariadbDialog';
import {
  getMongoDBValidationSchema,
  MongoDBDialog,
} from '../components/resources/dialogs/mongoDbDialog';
import {
  getMySQLValidationSchema,
  MysqlDialog,
} from '../components/resources/dialogs/mysqlDialog';
import OnDemandKubernetesDialog, {
  getOnDemandKubernetesValidationSchema,
} from '../components/resources/dialogs/onDemandKubernetesDialog';
import {
  getPostgresValidationSchema,
  PostgresDialog,
} from '../components/resources/dialogs/postgresDialog';
import {
  getRedshiftValidationSchema,
  RedshiftDialog,
} from '../components/resources/dialogs/redshiftDialog';
import {
  getS3ValidationSchema,
  S3Dialog,
} from '../components/resources/dialogs/s3Dialog';
import {
  getSlackValidationSchema,
  SlackDialog,
} from '../components/resources/dialogs/slackDialog';
import {
  getSnowflakeValidationSchema,
  SnowflakeDialog,
} from '../components/resources/dialogs/snowflakeDialog';
import {
  getSparkValidationSchema,
  SparkDialog,
} from '../components/resources/dialogs/sparkDialog';
import {
  getSQLiteValidationSchema,
  SQLiteDialog,
} from '../components/resources/dialogs/sqliteDialog';
import { AqueductDocsLink } from './docs';
import {
  AirflowConfig,
  AthenaConfig,
  AWSConfig,
  BigQueryConfig,
  DatabricksConfig,
  ECRConfig,
  EmailConfig,
  GarConfig,
  GCSConfig,
  LambdaConfig,
  MariaDbConfig,
  MongoDBConfig,
  MySqlConfig,
  PostgresConfig,
  RedshiftConfig,
  ResourceCategories,
  S3Config,
  ServiceInfoMap,
  ServiceLogos,
  SlackConfig,
  SnowflakeConfig,
  SparkConfig,
  SQLiteConfig,
} from './resources';

const addingResourceLink = `${AqueductDocsLink}/resources/adding-an-resource`;

export const SupportedResources: ServiceInfoMap = {
  ['Postgres']: {
    logo: ServiceLogos['Postgres'],
    activated: true,
    category: ResourceCategories.DATA,
    docs: addingResourceLink,
    dialog: ({ user, resourceToEdit, onCloseDialog, loading, disabled }) => (
      <PostgresDialog
        user={user}
        resourceToEdit={resourceToEdit as PostgresConfig}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: (editMode) => getPostgresValidationSchema(editMode),
  },
  ['Snowflake']: {
    logo: ServiceLogos['Snowflake'],
    activated: true,
    category: ResourceCategories.DATA,
    docs: addingResourceLink,
    dialog: ({ user, resourceToEdit, onCloseDialog, loading, disabled }) => (
      <SnowflakeDialog
        user={user}
        resourceToEdit={resourceToEdit as SnowflakeConfig}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: (editMode) => getSnowflakeValidationSchema(editMode),
  },
  ['Redshift']: {
    logo: ServiceLogos['Redshift'],
    activated: true,
    category: ResourceCategories.DATA,
    docs: addingResourceLink,
    dialog: ({ user, resourceToEdit, onCloseDialog, loading, disabled }) => (
      <RedshiftDialog
        user={user}
        resourceToEdit={resourceToEdit as RedshiftConfig}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: (editMode) => getRedshiftValidationSchema(editMode),
  },
  ['BigQuery']: {
    logo: ServiceLogos['BigQuery'],
    activated: true,
    category: ResourceCategories.DATA,
    docs: `${addingResourceLink}/connecting-to-google-bigquery`,
    dialog: ({ user, resourceToEdit, onCloseDialog, loading, disabled }) => (
      <BigQueryDialog
        user={user}
        resourceToEdit={resourceToEdit as BigQueryConfig}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: (editMode) => getBigQueryValidationSchema(editMode),
  },
  ['MySQL']: {
    logo: ServiceLogos['MySQL'],
    activated: true,
    category: ResourceCategories.DATA,
    docs: addingResourceLink,
    dialog: ({ user, resourceToEdit, onCloseDialog, loading, disabled }) => (
      <MysqlDialog
        user={user}
        resourceToEdit={resourceToEdit as MySqlConfig}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: (editMode) => getMySQLValidationSchema(editMode),
  },
  ['MariaDB']: {
    logo: ServiceLogos['MariaDB'],
    activated: true,
    category: ResourceCategories.DATA,
    docs: addingResourceLink,
    dialog: ({ user, resourceToEdit, onCloseDialog, loading, disabled }) => (
      <MariaDbDialog
        user={user}
        resourceToEdit={resourceToEdit as MariaDbConfig}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: (editMode) => getMariaDBValidationSchema(editMode),
  },
  ['S3']: {
    logo: ServiceLogos['S3'],
    activated: true,
    category: ResourceCategories.DATA,
    docs: `${addingResourceLink}/connecting-to-aws-s3`,
    dialog: ({
      user,
      resourceToEdit,
      onCloseDialog,
      loading,
      disabled,
      setMigrateStorage,
    }) => (
      <S3Dialog
        user={user}
        resourceToEdit={resourceToEdit as S3Config}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
        setMigrateStorage={setMigrateStorage}
      />
    ),
    validationSchema: (editMode) => getS3ValidationSchema(editMode),
  },
  ['GCS']: {
    logo: ServiceLogos['GCS'],
    activated: true,
    category: ResourceCategories.DATA,
    docs: `${addingResourceLink}/connecting-to-google-cloud-storage`,
    dialog: ({
      user,
      resourceToEdit,
      onCloseDialog,
      loading,
      disabled,
      setMigrateStorage,
    }) => (
      <GCSDialog
        user={user}
        resourceToEdit={resourceToEdit as GCSConfig}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
        setMigrateStorage={setMigrateStorage}
      />
    ),
    validationSchema: (editMode) => getGCSValidationSchema(editMode),
  },
  ['Aqueduct']: {
    logo: ServiceLogos['Aqueduct'],
    activated: true,
    category: ResourceCategories.COMPUTE,
    docs: addingResourceLink,
    dialog: ({}) => null,
    validationSchema: null,
  },
  ['Filesystem']: {
    logo: ServiceLogos['Aqueduct'],
    activated: true,
    category: ResourceCategories.ARTIFACT_STORAGE,
    docs: addingResourceLink,
    dialog: ({}) => null,
    validationSchema: null,
  },
  ['SQLite']: {
    logo: ServiceLogos['SQLite'],
    activated: true,
    category: ResourceCategories.DATA,
    docs: addingResourceLink,
    dialog: ({ user, resourceToEdit, onCloseDialog, loading, disabled }) => (
      <SQLiteDialog
        user={user}
        resourceToEdit={resourceToEdit as SQLiteConfig}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: ({}) => getSQLiteValidationSchema(),
  },
  ['Athena']: {
    logo: ServiceLogos['Athena'],
    activated: true,
    category: ResourceCategories.DATA,
    docs: addingResourceLink,
    dialog: ({ user, resourceToEdit, onCloseDialog, loading, disabled }) => (
      <AthenaDialog
        user={user}
        resourceToEdit={resourceToEdit as AthenaConfig}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: (editMode) => getAthenaValidationSchema(editMode),
  },
  ['Airflow']: {
    logo: ServiceLogos['Airflow'],
    activated: true,
    category: ResourceCategories.COMPUTE,
    docs: addingResourceLink,
    dialog: ({ user, resourceToEdit, onCloseDialog, loading, disabled }) => (
      <AirflowDialog
        user={user}
        resourceToEdit={resourceToEdit as AirflowConfig}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: (editMode) => getAirflowValidationSchema(editMode),
  },
  ['Kubernetes']: {
    logo: ServiceLogos['Kubernetes'],
    activated: true,
    category: ResourceCategories.COMPUTE,
    docs: addingResourceLink,
    dialog: ({ user, resourceToEdit, onCloseDialog, loading, disabled }) => (
      <OnDemandKubernetesDialog
        user={user}
        resourceToEdit={resourceToEdit}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: (editMode) =>
      getOnDemandKubernetesValidationSchema(editMode),
  },
  ['Lambda']: {
    logo: ServiceLogos['Lambda'],
    activated: true,
    category: ResourceCategories.COMPUTE,
    docs: `${addingResourceLink}/connecting-to-aws-lambda`,
    dialog: ({ user, resourceToEdit, onCloseDialog, loading, disabled }) => (
      <LambdaDialog
        user={user}
        resourceToEdit={resourceToEdit as LambdaConfig}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: ({}) => getLambdaValidationSchema(),
  },
  ['MongoDB']: {
    logo: ServiceLogos['MongoDB'],
    activated: true,
    category: ResourceCategories.DATA,
    docs: addingResourceLink,
    dialog: ({ user, resourceToEdit, onCloseDialog, loading, disabled }) => (
      <MongoDBDialog
        user={user}
        resourceToEdit={resourceToEdit as MongoDBConfig}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: (editMode) => getMongoDBValidationSchema(editMode),
  },
  ['Conda']: {
    logo: ServiceLogos['Conda'],
    activated: true,
    category: ResourceCategories.COMPUTE,
    docs: `${addingResourceLink}/connecting-to-conda`,
    dialog: ({ user, onCloseDialog, loading, disabled }) => (
      <CondaDialog
        user={user}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: ({}) => getCondaValidationSchema(),
  },
  ['Databricks']: {
    logo: ServiceLogos['Databricks'],
    activated: true,
    category: ResourceCategories.COMPUTE,
    docs: `${addingResourceLink}/connecting-to-databricks`,
    dialog: ({ user, resourceToEdit, onCloseDialog, loading, disabled }) => (
      <DatabricksDialog
        user={user}
        resourceToEdit={resourceToEdit as DatabricksConfig}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: (editMode) => getDatabricksValidationSchema(editMode),
  },
  ['Email']: {
    logo: ServiceLogos['Email'],
    activated: true,
    category: ResourceCategories.NOTIFICATION,
    docs: `${AqueductDocsLink}/notifications/connecting-to-email`,
    dialog: ({ user, resourceToEdit, onCloseDialog, loading, disabled }) => (
      <EmailDialog
        user={user}
        resourceToEdit={resourceToEdit as EmailConfig}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: (editMode) => getEmailValidationSchema(editMode),
  },
  ['Slack']: {
    logo: ServiceLogos['Slack'],
    activated: true,
    category: ResourceCategories.NOTIFICATION,
    docs: `${AqueductDocsLink}/notifications/connecting-to-slack`,
    dialog: ({ user, resourceToEdit, onCloseDialog, loading, disabled }) => (
      <SlackDialog
        user={user}
        resourceToEdit={resourceToEdit as SlackConfig}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: (editMode) => getSlackValidationSchema(editMode),
  },
  ['Spark']: {
    logo: ServiceLogos['Spark'],
    activated: true,
    category: ResourceCategories.COMPUTE,
    docs: addingResourceLink,
    dialog: ({ user, resourceToEdit, onCloseDialog, loading, disabled }) => (
      <SparkDialog
        user={user}
        resourceToEdit={resourceToEdit as SparkConfig}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: ({}) => getSparkValidationSchema(),
  },
  // Not sure the difference between this one and the Amazon one below.
  ['AWS']: {
    logo: ServiceLogos['AWS'],
    activated: true,
    category: ResourceCategories.CLOUD,
    docs: addingResourceLink,
    dialog: ({ user, resourceToEdit, onCloseDialog, loading, disabled }) => (
      <AWSDialog
        user={user}
        resourceToEdit={resourceToEdit as AWSConfig}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: (editMode) => getAWSValidationSchema(editMode),
  },
  ['Amazon']: {
    logo: ServiceLogos['AWS'],
    activated: true,
    category: ResourceCategories.CLOUD,
    docs: addingResourceLink,
    dialog: ({ user, resourceToEdit, onCloseDialog, loading, disabled }) => (
      <AWSDialog
        user={user}
        resourceToEdit={resourceToEdit as AWSConfig}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: (editMode) => getAWSValidationSchema(editMode),
  },
  ['GCP']: {
    logo: ServiceLogos['GCP'],
    activated: false,
    category: ResourceCategories.CLOUD,
    docs: addingResourceLink,
    dialog: ({}) => <GCPDialog />,
    validationSchema: ({}) => getGCPValidationSchema(),
  },
  ['Azure']: {
    logo: ServiceLogos['Azure'],
    activated: false,
    category: ResourceCategories.CLOUD,
    docs: addingResourceLink,
    dialog: ({}) => <AzureDialog />,
    validationSchema: ({}) => getAzureValidationSchema(),
  },
  ['ECR']: {
    logo: ServiceLogos['ECR'],
    activated: true,
    category: ResourceCategories.CONTAINER_REGISTRY,
    docs: addingResourceLink,
    dialog: ({ user, resourceToEdit, onCloseDialog, loading, disabled }) => (
      <ECRDialog
        user={user}
        resourceToEdit={resourceToEdit as ECRConfig}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: (editMode) => getECRValidationSchema(editMode),
  },
  ['GAR']: {
    logo: ServiceLogos['GAR'],
    activated: true,
    category: ResourceCategories.CONTAINER_REGISTRY,
    docs: addingResourceLink,
    dialog: ({ user, resourceToEdit, onCloseDialog, loading, disabled }) => (
      <GARDialog
        user={user}
        resourceToEdit={resourceToEdit as GarConfig}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
      />
    ),
    validationSchema: (editMode) => getGARValidationSchema(editMode),
  },
};

export default SupportedResources;
