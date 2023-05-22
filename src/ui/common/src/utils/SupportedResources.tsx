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
import { ResourceCategories, ServiceInfoMap, ServiceLogos } from './resources';

const addingResourceLink = `${AqueductDocsLink}/resources/adding-an-resource`;

export const SupportedResources: ServiceInfoMap = {
  ['Postgres']: {
    logo: ServiceLogos['Postgres'],
    activated: true,
    category: ResourceCategories.DATA,
    docs: addingResourceLink,
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
    category: ResourceCategories.DATA,
    docs: addingResourceLink,
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
    category: ResourceCategories.DATA,
    docs: addingResourceLink,
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
    category: ResourceCategories.DATA,
    docs: `${addingResourceLink}/connecting-to-google-bigquery`,
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
    category: ResourceCategories.DATA,
    docs: addingResourceLink,
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
    category: ResourceCategories.DATA,
    docs: addingResourceLink,
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
    category: ResourceCategories.DATA,
    docs: `${addingResourceLink}/connecting-to-aws-s3`,
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
    category: ResourceCategories.DATA,
    docs: `${addingResourceLink}/connecting-to-google-cloud-storage`,
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
    category: ResourceCategories.COMPUTE,
    docs: addingResourceLink,
    // TODO: Figure out what to show here.
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => <div />,
    validationSchema: null,
  },
  ['Filesystem']: {
    logo: ServiceLogos['Aqueduct'],
    activated: true,
    category: ResourceCategories.ARTIFACT_STORAGE,
    docs: addingResourceLink,
    dialog: ({ user, editMode, onCloseDialog, loading, disabled }) => null,
    validationSchema: getSQLiteValidationSchema(),
  },
  ['SQLite']: {
    logo: ServiceLogos['SQLite'],
    activated: true,
    category: ResourceCategories.DATA,
    docs: addingResourceLink,
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
    category: ResourceCategories.DATA,
    docs: addingResourceLink,
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
    category: ResourceCategories.COMPUTE,
    docs: addingResourceLink,
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
    category: ResourceCategories.COMPUTE,
    docs: addingResourceLink,
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
  //   category: ResourceCategories.COMPUTE,
  //   docs: `${addingResourceLink}/connecting-to-k8s-cluster`,
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
    category: ResourceCategories.COMPUTE,
    docs: `${addingResourceLink}/connecting-to-aws-lambda`,
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
    category: ResourceCategories.DATA,
    docs: addingResourceLink,
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
    category: ResourceCategories.COMPUTE,
    docs: `${addingResourceLink}/connecting-to-conda`,
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
    category: ResourceCategories.COMPUTE,
    docs: `${addingResourceLink}/connecting-to-databricks`,
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
    category: ResourceCategories.NOTIFICATION,
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
    category: ResourceCategories.NOTIFICATION,
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
    category: ResourceCategories.COMPUTE,
    docs: addingResourceLink,
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
    category: ResourceCategories.CLOUD,
    docs: addingResourceLink,
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
    category: ResourceCategories.CLOUD,
    docs: addingResourceLink,
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
    category: ResourceCategories.CLOUD,
    docs: addingResourceLink,
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
    category: ResourceCategories.CLOUD,
    docs: addingResourceLink,
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
    category: ResourceCategories.CONTAINER_REGISTRY,
    docs: addingResourceLink,
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

export default SupportedResources;
