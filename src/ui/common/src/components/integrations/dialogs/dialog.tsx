import { LoadingButton } from '@mui/lab';
import {
  Alert,
  AlertTitle,
  Box,
  DialogActions,
  DialogContent,
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
  BigQueryConfig,
  CSVConfig,
  formatService,
  Integration,
  IntegrationConfig,
  MySqlConfig,
  PostgresConfig,
  RedshiftConfig,
  S3Config,
  Service,
  SnowflakeConfig,
  SupportedIntegrations,
} from '../../../utils/integrations';
import { isFailed, isLoading, isSucceeded } from '../../../utils/shared';
import { AirflowDialog } from './airflowDialog';
import { BigQueryDialog } from './bigqueryDialog';
import { IntegrationTextInputField } from './IntegrationTextInputField';
import { MariaDbDialog } from './mariadbDialog';
import { MysqlDialog } from './mysqlDialog';
import { PostgresDialog } from './postgresDialog';
import { RedshiftDialog } from './redshiftDialog';
import { isS3ConfigComplete, S3Dialog } from './s3Dialog';
import { SnowflakeDialog } from './snowflakeDialog';

type Props = {
  user: UserProfile;
  service: Service;
  onCloseDialog: () => void;
  onSuccess: () => void;
  integrationToEdit?: Integration;
};

const IntegrationDialog: React.FC<Props> = ({
  user,
  service,
  onCloseDialog,
  onSuccess,
  integrationToEdit = undefined,
}) => {
  const editMode = !!integrationToEdit;
  const dispatch: AppDispatch = useDispatch();
  const [config, setConfig] = useState<IntegrationConfig>(
    editMode
      ? { ...integrationToEdit.config } // make a copy to avoid accessing a state object
      : {}
  );
  const [name, setName] = useState<string>(
    editMode ? integrationToEdit.name : ''
  );

  const connectNewStatus = useSelector(
    (state: RootState) => state.integrationReducer.connectNewStatus
  );

  const editStatus = useSelector(
    (state: RootState) => state.integrationReducer.editStatus
  );

  const operators = useSelector(
    (state: RootState) => state.integrationReducer.operators.operators
  );

  const numWorkflows = new Set(operators.map((x) => x.workflow_id)).size;

  const connectStatus = editMode ? editStatus : connectNewStatus;
  const disableConnect =
    !editMode &&
    (!isConfigComplete(config) || name === '' || name === aqueductDemoName);
  const setConfigField = (field: string, value: string) =>
    setConfig((config) => {
      return { ...config, [field]: value };
    });

  useEffect(() => {
    if (isSucceeded(connectStatus)) {
      dispatch(
        handleLoadIntegrations({ apiKey: user.apiKey, forceLoad: true })
      );
      onSuccess();
      onCloseDialog();
    }
  }, [connectStatus]);

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

  let serviceDialog;

  switch (service) {
    case 'Postgres':
      serviceDialog = (
        <PostgresDialog
          onUpdateField={setConfigField}
          value={config as PostgresConfig}
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
        />
      );
      break;
    case 'Redshift':
      serviceDialog = (
        <RedshiftDialog
          onUpdateField={setConfigField}
          value={config as RedshiftConfig}
        />
      );
      break;
    case 'MariaDB':
      serviceDialog = (
        <MariaDbDialog
          onUpdateField={setConfigField}
          value={config as RedshiftConfig}
        />
      );
      break;
    case 'BigQuery':
      serviceDialog = (
        <BigQueryDialog
          onUpdateField={setConfigField}
          value={config as BigQueryConfig}
        />
      );
      break;
    case 'S3':
      serviceDialog = (
        <S3Dialog onUpdateField={setConfigField} value={config as S3Config} />
      );
      break;
    case 'Airflow':
      serviceDialog = (
        <AirflowDialog
          onUpdateField={setConfigField}
          value={config as AirflowConfig}
        />
      );
      break;
    default:
      return null;
  }

  const onConfirmDialog = () => {
    editMode
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
            {`The update will take effect for all ${numWorkflows} ${
              numWorkflows === 1 ? 'workflow' : 'workflows'
            } using this integration.`}
          </Alert>
        )}
        {nameInput}
        {serviceDialog}

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

// Helper function to check if the Integration config is completely filled
export function isConfigComplete(
  config: IntegrationConfig | CSVConfig
): boolean {
  if (isS3ConfigComplete(config as S3Config)) {
    return true;
  }

  return (
    Object.values(config).length > 0 &&
    Object.values(config).every((x) => x === undefined || (x && x !== ''))
  );
}

export default IntegrationDialog;
