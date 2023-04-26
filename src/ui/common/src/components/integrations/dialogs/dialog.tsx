import { DevTool } from '@hookform/devtools';
import { yupResolver } from '@hookform/resolvers/yup';
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
import { FormProvider, useForm } from 'react-hook-form';
import { useDispatch, useSelector } from 'react-redux';
import * as Yup from 'yup';

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
import { isAirflowConfigComplete } from './airflowDialog';
import { isAthenaConfigComplete } from './athenaDialog';
import { isAWSConfigComplete } from './awsDialog';
import { isBigQueryConfigComplete } from './bigqueryDialog';
import { isDatabricksConfigComplete } from './databricksDialog';
import { EmailDefaultsOnCreate, isEmailConfigComplete } from './emailDialog';
import { isGCSConfigComplete } from './gcsDialog';
import { IntegrationTextInputField } from './IntegrationTextInputField';
import { isK8sConfigComplete } from './kubernetesDialog';
import { isLambaDialogComplete } from './lambdaDialog';
import { isMariaDBConfigComplete } from './mariadbDialog';
import { isMongoDBConfigComplete } from './mongoDbDialog';
import { isMySqlConfigComplete } from './mysqlDialog';
import { isPostgresConfigComplete } from './postgresDialog';
import { isRedshiftConfigComplete } from './redshiftDialog';
import { isS3ConfigComplete } from './s3Dialog';
import { isSlackConfigComplete, SlackDefaultsOnCreate } from './slackDialog';
import { isSnowflakeConfigComplete } from './snowflakeDialog';
import { isSparkConfigComplete } from './sparkDialog';
import { isSQLiteConfigComplete } from './sqliteDialog';

type Props = {
  user: UserProfile;
  service: Service;
  onCloseDialog: () => void;
  onSuccess: () => void;
  showMigrationDialog?: () => void;
  integrationToEdit?: Integration;
  dialogContent: React.FC;
  validationSchema: Yup.ObjectSchema<any>;
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
  dialogContent,
  validationSchema,
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

  const numWorkflows = operators
    ? new Set(operators.map((x) => x.workflow_id)).size
    : 0;

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

  // TODO: Figure out how we're going to set up validation schema
  //const { register, control, handleSubmit, formState } = useForm();

  // How do i use Yup.inferType to get the type of the validationSchema?
  // Yup.inferType<typeof validationSchema>;
  //type Inferred = Yup.InferType<typeof validationSchema>;

  // const validationSchema = Yup.object().shape({
  //   name: Yup.string().required('Please enter a name.'),
  //   host: Yup.string().required('Please enter a host url.'),
  //   port: Yup.string().required('Please enter a port number.'),
  //   //database: Yup.string().required('Please enter a database name.'),
  //   //username: Yup.string().required('Please enter a username.'),
  //   //password: Yup.string().required('Please enter a password.'),
  // });

  console.log('validationSchema: ', validationSchema);

  const methods = useForm({
    // TODO: Figure out how to get the validationSchema from the appropriate dialog.
    resolver: yupResolver(validationSchema),
  });

  const onSubmit = (data: any) => {
    console.log('inside onSubmit');
    console.log(JSON.stringify(data, null, 2));
  };

  console.log('formState from Dialog.tsx: ', methods.formState);
  console.log('Dialog.tsx touchedFields: ', methods.formState.touchedFields);
  console.log('Dialog.tsx errors: ', methods.formState.touchedFields);
  console.log('Dialog.tsx getValues: ', methods.getValues());

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
          : `Connecting to ${service}`}
      </Typography>
      <img height="45px" src={SupportedIntegrations[service].logo} />
    </Box>
  );

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
      name="name"
      spellCheck={false}
      required={true}
      label="Name*"
      description="Provide a unique name to refer to this integration."
      placeholder={'my_' + formatService(service) + '_integration'}
      onChange={(event) => {
        setShouldShowNameError(false);
        methods.setValue('name', event.target.value);
      }}
      disabled={service === 'Aqueduct Demo'}
    />
  );

  return (
    <Dialog open={true} onClose={onCloseDialog} fullWidth maxWidth="lg">
      <FormProvider {...methods}>
        <form>
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
                <Link
                  href={SupportedIntegrations[service].docs}
                  target="_blank"
                >
                  documentation
                </Link>
                .
              </Typography>
            )}
            {nameInput}

            {dialogContent({ editMode })}

            {shouldShowNameError && (
              <Alert sx={{ mt: 2 }} severity="error">
                <AlertTitle>Naming Error</AlertTitle>A connected integration
                already exists with this name. Please provide a unique name for
                your integration.
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
              onClick={async () => {
                console.log('loading button clicked. Calling handleSubmit()');
                console.log('formState: ', methods.formState);

                const triggerResult = await methods.trigger();
                console.log('triggerResult: ', triggerResult);

                const triggerParams = await methods.trigger([
                  'name',
                  'host',
                  'port',
                ]);

                console.log('triggerParams: ', triggerParams);

                // NOTE: handleSubmit() is a function that returns a function, please call it as so
                methods.handleSubmit(onSubmit)();
              }}
              loading={isLoading(connectStatus)}
              //disabled={disableConnect}
              disabled={false}
            >
              Confirm
            </LoadingButton>
          </DialogActions>
        </form>
      </FormProvider>
      <DevTool control={methods.control} /> {/* set up the dev tool */}
    </Dialog>
  );
};

// TODO: refactor this so that we no longer need a switch statement here.
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
    case 'BigQuery':
      return isBigQueryConfigComplete(config as BigQueryConfig);
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
      return isLambaDialogComplete(config as LambdaConfig);
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
    default:
      // Require all integrations to have their own validation function.
      return false;
  }
}

export default IntegrationDialog;
