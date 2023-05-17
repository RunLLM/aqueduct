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
import {
  CondaDialog,
  getCondaValidationSchema,
} from '../components/integrations/dialogs/condaDialog';
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
    validationSchema: getCondaValidationSchema(),
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
