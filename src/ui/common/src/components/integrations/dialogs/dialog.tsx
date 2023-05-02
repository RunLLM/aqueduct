import { LoadingButton } from '@mui/lab';
import {
  Alert,
  AlertTitle,
  Box,
  DialogActions,
  DialogContent,
  Link,
  Typography,
} from '@mui/material';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import {
  handleConnectToNewIntegration,
  handleEditIntegration,
} from '../../../reducers/integration';
import { handleLoadIntegrations } from '../../../reducers/integrations';
import { AppDispatch, RootState } from '../../../stores/store';
import UserProfile from '../../../utils/auth';
import {
  AirflowConfig,
  aqueductDemoName,
  AthenaConfig,
  AWSConfig,
  BigQueryConfig,
  DatabricksConfig,
  ECRConfig,
  EmailConfig,
  formatService,
  GCSConfig,
  Integration,
  IntegrationConfig,
  KubernetesConfig,
  LambdaConfig,
  MariaDbConfig,
  MongoDBConfig,
  MySqlConfig,
  PostgresConfig,
  RedshiftConfig,
  S3Config,
  Service,
  SlackConfig,
  SnowflakeConfig,
  SparkConfig,
  SQLiteConfig,
  SupportedIntegrations,
} from '../../../utils/integrations';
import { isFailed, isLoading, isSucceeded } from '../../../utils/shared';
import { AirflowDialog, isAirflowConfigComplete } from './airflowDialog';
import { AthenaDialog, isAthenaConfigComplete } from './athenaDialog';
import { AWSDialog, isAWSConfigComplete } from './awsDialog';
import { BigQueryDialog } from './bigqueryDialog';
import { CondaDialog } from './condaDialog';
import {
  DatabricksDialog,
  isDatabricksConfigComplete,
} from './databricksDialog';
import { ECRDialog, isECRConfigComplete } from './ecrDialog';
import {
  EmailDefaultsOnCreate,
  EmailDialog,
  isEmailConfigComplete,
} from './emailDialog';
import { GCSDialog, isGCSConfigComplete } from './gcsDialog';
import { IntegrationTextInputField } from './IntegrationTextInputField';
import { isK8sConfigComplete, KubernetesDialog } from './kubernetesDialog';
import { LambdaDialog } from './lambdaDialog';
import { isMariaDBConfigComplete, MariaDbDialog } from './mariadbDialog';
import { isMongoDBConfigComplete, MongoDBDialog } from './mongoDbDialog';
import { isMySqlConfigComplete, MysqlDialog } from './mysqlDialog';
import { isPostgresConfigComplete, PostgresDialog } from './postgresDialog';
import { isRedshiftConfigComplete, RedshiftDialog } from './redshiftDialog';
import { isS3ConfigComplete, S3Dialog } from './s3Dialog';
import {
  isSlackConfigComplete,
  SlackDefaultsOnCreate,
  SlackDialog,
} from './slackDialog';
import { isSnowflakeConfigComplete, SnowflakeDialog } from './snowflakeDialog';
import { isSparkConfigComplete, SparkDialog } from './sparkDialog';
import { isSQLiteConfigComplete, SQLiteDialog } from './sqliteDialog';

type Props = {
  user: UserProfile;
  service: Service;
  onCloseDialog: () => void;
  onSuccess: () => void;
  showMigrationDialog?: () => void;
  integrationToEdit?: Integration;
};

// Default fields are actual filled form values on 'create' dialog.
function defaultFields(service: Service): IntegrationConfig {
  switch (service) {
    case 'Email':
      return EmailDefaultsOnCreate as EmailConfig;

    case 'Slack':
      return SlackDefaultsOnCreate as SlackConfig;
  }

  return {};
}

const IntegrationDialog: React.FC<Props> = ({
  user,
  service,
  onCloseDialog,
  onSuccess,
  showMigrationDialog = undefined,
  integrationToEdit = undefined,
}) => {
  const editMode = !!integrationToEdit;
  const dispatch: AppDispatch = useDispatch();
  const [config, setConfig] = useState<IntegrationConfig>(
    editMode
      ? { ...integrationToEdit.config } // make a copy to avoid accessing a state object
      : { ...defaultFields(service) }
  );
  const [name, setName] = useState<string>(
    editMode ? integrationToEdit.name : ''
  );

  const [shouldShowNameError, setShouldShowNameError] =
    useState<boolean>(false);

  const connectNewStatus = useSelector(
    (state: RootState) => state.integrationReducer.connectNewStatus
  );

  const editStatus = useSelector(
    (state: RootState) => state.integrationReducer.editStatus
  );

  const operators = useSelector(
    (state: RootState) => state.integrationReducer.operators.operators
  );

  const integrations = useSelector((state: RootState) =>
    Object.values(state.integrationsReducer.integrations)
  );

  const numWorkflows = new Set(operators.map((x) => x.workflow_id)).size;

  const connectStatus = editMode ? editStatus : connectNewStatus;
  const disableConnect =
    !editMode &&
    (!isConfigComplete(config, service) ||
      name === '' ||
      name === aqueductDemoName);
  const setConfigField = (field: string, value: string) =>
    setConfig((config) => {
      return { ...config, [field]: value };
    });

  const [migrateStorage, setMigrateStorage] = useState(false);

  useEffect(() => {
    if (isSucceeded(connectStatus)) {
      dispatch(
        handleLoadIntegrations({ apiKey: user.apiKey, forceLoad: true })
      );
      onSuccess();
      if (showMigrationDialog && migrateStorage) {
        showMigrationDialog();
      }
      onCloseDialog();
    }
  }, [
    connectStatus,
    dispatch,
    migrateStorage,
    onCloseDialog,
    onSuccess,
    showMigrationDialog,
    user.apiKey,
  ]);

  let connectionMessage = '';
  if (service === 'AWS') {
    connectionMessage = 'Configuring Aqueduct-managed Kubernetes on AWS';
  } else {
    connectionMessage = `Connecting to ${service}`;
  }

  const dialogHeader = (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'row',
        justifyContent: 'space-between',
        width: '100%',
      }}
    >
      <Typography variant="h5">
        {!!integrationToEdit
          ? `Edit ${integrationToEdit.name}`
          : `${connectionMessage}`}
      </Typography>
      <img height="45px" src={SupportedIntegrations[service].logo} />
    </Box>
  );

  let serviceDialog;

  switch (service) {
    case 'Postgres':
      serviceDialog = (
        <PostgresDialog
          onUpdateField={setConfigField}
          value={config as PostgresConfig}
          editMode={editMode}
        />
      );
      break;
    case 'Snowflake':
      serviceDialog = (
        <SnowflakeDialog
          onUpdateField={setConfigField}
          value={config as SnowflakeConfig}
          editMode={editMode}
        />
      );
      break;
    case 'Aqueduct Demo':
      serviceDialog = null;
      break;
    case 'MySQL':
      serviceDialog = (
        <MysqlDialog
          onUpdateField={setConfigField}
          value={config as MySqlConfig}
          editMode={editMode}
        />
      );
      break;
    case 'Redshift':
      serviceDialog = (
        <RedshiftDialog
          onUpdateField={setConfigField}
          value={config as RedshiftConfig}
          editMode={editMode}
        />
      );
      break;
    case 'MariaDB':
      serviceDialog = (
        <MariaDbDialog
          onUpdateField={setConfigField}
          value={config as MariaDbConfig}
          editMode={editMode}
        />
      );
      break;
    case 'MongoDB':
      serviceDialog = (
        <MongoDBDialog
          onUpdateField={setConfigField}
          value={config as MongoDBConfig}
          editMode={editMode}
        />
      );
      break;
    case 'BigQuery':
      serviceDialog = (
        <BigQueryDialog
          onUpdateField={setConfigField}
          value={config as BigQueryConfig}
          editMode={editMode}
        />
      );
      break;
    case 'S3':
      serviceDialog = (
        <S3Dialog
          onUpdateField={setConfigField}
          value={config as S3Config}
          editMode={editMode}
          setMigrateStorage={setMigrateStorage}
        />
      );
      break;
    case 'GCS':
      const gcsConfig = config as GCSConfig;
      // GCS can only be used storage currently
      gcsConfig.use_as_storage = 'true';
      serviceDialog = (
        <GCSDialog
          onUpdateField={setConfigField}
          value={config as GCSConfig}
          editMode={editMode}
          setMigrateStorage={setMigrateStorage}
        />
      );
      break;
    case 'Athena':
      serviceDialog = (
        <AthenaDialog
          onUpdateField={setConfigField}
          value={config as AthenaConfig}
          editMode={editMode}
        />
      );
      break;
    case 'Airflow':
      serviceDialog = (
        <AirflowDialog
          onUpdateField={setConfigField}
          value={config as AirflowConfig}
          editMode={editMode}
        />
      );
      break;
    case 'Kubernetes':
      serviceDialog = (
        <KubernetesDialog
          onUpdateField={setConfigField}
          value={config as KubernetesConfig}
          apiKey={user.apiKey}
        />
      );
      break;
    case 'Lambda':
      serviceDialog = (
        <LambdaDialog
          onUpdateField={setConfigField}
          value={config as LambdaConfig}
        />
      );
      break;
    case 'SQLite':
      serviceDialog = (
        <SQLiteDialog
          onUpdateField={setConfigField}
          value={config as SQLiteConfig}
          editMode={editMode}
        />
      );
      break;
    case 'Conda':
      serviceDialog = <CondaDialog />;
      break;
    case 'Databricks':
      serviceDialog = (
        <DatabricksDialog
          onUpdateField={setConfigField}
          value={config as DatabricksConfig}
          editMode={editMode}
        />
      );
      break;
    case 'Email':
      serviceDialog = (
        <EmailDialog
          onUpdateField={setConfigField}
          value={config as EmailConfig}
        />
      );
      break;
    case 'Slack':
      serviceDialog = (
        <SlackDialog
          onUpdateField={setConfigField}
          value={config as SlackConfig}
        />
      );
      break;
    case 'Spark':
      serviceDialog = (
        <SparkDialog
          onUpdateField={setConfigField}
          value={config as SparkConfig}
          editMode={editMode}
        />
      );
      break;
    case 'AWS':
      serviceDialog = (
        <AWSDialog onUpdateField={setConfigField} value={config as AWSConfig} />
      );
      break;
    case 'ECR':
      serviceDialog = (
        <ECRDialog onUpdateField={setConfigField} value={config as AWSConfig} />
      );
      break;
    default:
      return null;
  }

  const onConfirmDialog = () => {
    //check that name is unique before connecting.
    if (!editMode) {
      for (let i = 0; i < integrations.length; i++) {
        if (name === integrations[i].name) {
          setShouldShowNameError(true);
          return;
        }
      }
    }

    return editMode
      ? dispatch(
          handleEditIntegration({
            apiKey: user.apiKey,
            integrationId: integrationToEdit.id,
            name: name,
            config: config,
          })
        )
      : dispatch(
          handleConnectToNewIntegration({
            apiKey: user.apiKey,
            service: service,
            name: name,
            config: config,
          })
        );
  };

  const nameInput = (
    <IntegrationTextInputField
      spellCheck={false}
      required={true}
      label="Name*"
      description="Provide a unique name to refer to this integration."
      placeholder={'my_' + formatService(service) + '_integration'}
      onChange={(event) => {
        setName(event.target.value);
        setShouldShowNameError(false);
      }}
      value={name}
      disabled={service === 'Aqueduct Demo'}
    />
  );

  return (
    <Dialog open={true} onClose={onCloseDialog} fullWidth maxWidth="lg">
      <DialogTitle>{dialogHeader}</DialogTitle>
      <DialogContent>
        {editMode && numWorkflows > 0 && (
          <Alert sx={{ mb: 2 }} severity="info">
            {`Changing this integration will automatically update ${numWorkflows} ${
              numWorkflows === 1 ? 'workflow' : 'workflows'
            }.`}
          </Alert>
        )}
        {(service === 'Email' || service === 'Slack') && (
          <Typography variant="body1" color="gray.700">
            To learn more about how to set up {service}, see our{' '}
            <Link href={SupportedIntegrations[service].docs} target="_blank">
              documentation
            </Link>
            .
          </Typography>
        )}
        {nameInput}
        {serviceDialog}

        {shouldShowNameError && (
          <Alert sx={{ mt: 2 }} severity="error">
            <AlertTitle>Naming Error</AlertTitle>A connected integration already
            exists with this name. Please provide a unique name for your
            integration.
          </Alert>
        )}

        {isFailed(connectStatus) && (
          <Alert sx={{ mt: 2 }} severity="error">
            <AlertTitle>
              {editMode
                ? `Failed to update ${integrationToEdit.name}`
                : `Unable to connect to ${service}`}
            </AlertTitle>
            <pre>{connectStatus.err}</pre>
          </Alert>
        )}
      </DialogContent>
      <DialogActions>
        <Button autoFocus onClick={onCloseDialog}>
          Cancel
        </Button>
        <LoadingButton
          autoFocus
          onClick={onConfirmDialog}
          loading={isLoading(connectStatus)}
          disabled={disableConnect}
        >
          Confirm
        </LoadingButton>
      </DialogActions>
    </Dialog>
  );
};

// Helper function to check if the Integration config is completely filled.
export function isConfigComplete(
  config: IntegrationConfig,
  service: Service
): boolean {
  switch (service) {
    case 'Airflow':
      return isAirflowConfigComplete(config as AirflowConfig);
    case 'Athena':
      return isAthenaConfigComplete(config as AthenaConfig);
    case 'AWS':
      return isAWSConfigComplete(config as AWSConfig);
    case 'Conda':
      // Conda only has a name field that the user supplies, so this half of form is always valid.
      return true;
    case 'Databricks':
      return isDatabricksConfigComplete(config as DatabricksConfig);
    case 'Email':
      return isEmailConfigComplete(config as EmailConfig);
    case 'GCS':
      return isGCSConfigComplete(config as GCSConfig);
    case 'Kubernetes':
      return isK8sConfigComplete(config as KubernetesConfig);
    case 'Lambda':
      // Lambda only has a name field that the user supplies, so this half of form is always valid.
      return true;
    case 'MariaDB':
      return isMariaDBConfigComplete(config as MariaDbConfig);
    case 'MongoDB':
      return isMongoDBConfigComplete(config as MongoDBConfig);
    case 'MySQL':
      return isMySqlConfigComplete(config as MySqlConfig);
    case 'Postgres':
      return isPostgresConfigComplete(config as PostgresConfig);
    case 'Redshift':
      return isRedshiftConfigComplete(config as RedshiftConfig);
    case 'S3':
      return isS3ConfigComplete(config as S3Config);
    case 'Slack':
      return isSlackConfigComplete(config as SlackConfig);
    case 'Spark':
      return isSparkConfigComplete(config as SparkConfig);
    case 'Snowflake':
      return isSnowflakeConfigComplete(config as SnowflakeConfig);
    case 'SQLite':
      return isSQLiteConfigComplete(config as SQLiteConfig);
    case 'ECR':
      return isECRConfigComplete(config as ECRConfig);
    default:
      // Require all integrations to have their own validation function.
      return false;
  }
}

export default IntegrationDialog;
