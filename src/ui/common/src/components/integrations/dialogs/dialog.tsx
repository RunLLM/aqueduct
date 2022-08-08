import { LoadingButton } from '@mui/lab';
import {
  Alert,
  AlertTitle,
  Box,
  DialogActions,
  DialogContent,
  Snackbar,
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
  aqueductDemoName,
  CSVConfig,
  Integration,
  IntegrationConfig,
  Service,
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
import { S3Dialog } from './s3Dialog';
import { SnowflakeDialog } from './snowflakeDialog';

type Props = {
  user: UserProfile;
  service: Service;
  onCloseDialog: () => void;
  integrationToEdit?: Integration;
};

export const IntegrationDialog: React.FC<Props> = ({
  user,
  service,
  onCloseDialog,
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
  const [disableConnect, setDisableConnect] = useState(true);
  const [showSuccessToast, setShowSuccessToast] = useState(false);

  const handleSuccessToastClose = () => {
    setShowSuccessToast(false);
  };

  const connectNewStatus = useSelector(
    (state: RootState) => state.integrationReducer.connectNewStatus
  );

  const editStatus = useSelector(
    (state: RootState) => state.integrationReducer.editStatus
  );

  const connectStatus = editMode ? editStatus : connectNewStatus;

  useEffect(() => {
    setDisableConnect(
      service !== 'Aqueduct Demo' &&
        (!isConfigComplete(config) || name === '' || name === aqueductDemoName)
    );
  }, [config, name]);

  useEffect(() => {
    if (isSucceeded(connectStatus)) {
      dispatch(
        handleLoadIntegrations({ apiKey: user.apiKey, forceLoad: true })
      );
      setShowSuccessToast(true);
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
      serviceDialog = <PostgresDialog setDialogConfig={setConfig} />;
      break;
    case 'Snowflake':
      serviceDialog = <SnowflakeDialog setDialogConfig={setConfig} />;
      break;
    case 'Aqueduct Demo':
      serviceDialog = null;
      break;
    case 'MySQL':
      serviceDialog = <MysqlDialog setDialogConfig={setConfig} />;
      break;
    case 'Redshift':
      serviceDialog = <RedshiftDialog setDialogConfig={setConfig} />;
      break;
    case 'MariaDB':
      serviceDialog = <MariaDbDialog setDialogConfig={setConfig} />;
      break;
    case 'BigQuery':
      serviceDialog = <BigQueryDialog setDialogConfig={setConfig} />;
      break;
    case 'S3':
      serviceDialog = <S3Dialog setDialogConfig={setConfig} />;
      break;
    case 'Airflow':
      serviceDialog = <AirflowDialog setDialogConfig={setConfig} />;
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
        {nameInput}
        {serviceDialog}

        {isFailed(connectStatus) && (
          <Alert sx={{ mt: 2 }} severity="error">
            <AlertTitle>
              {editMode
                ? `Failed to upddate ${integrationToEdit.name}`
                : `Unable to connect to ${service}`}
            </AlertTitle>
            <pre>{connectStatus.err}</pre>
          </Alert>
        )}
        <Snackbar
          anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
          open={showSuccessToast}
          onClose={handleSuccessToastClose}
          key={'integrations-dialog-success-snackbar'}
          autoHideDuration={6000}
        >
          <Alert
            onClose={handleSuccessToastClose}
            severity="success"
            sx={{ width: '100%' }}
          >
            {editMode
              ? `Successfully updated ${integrationToEdit.name}!`
              : `Successfully connected to ${service}!`}
          </Alert>
        </Snackbar>
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
  return Object.values(config).every((x) => x === undefined || (x && x !== ''));
}
