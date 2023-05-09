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
import React, { MouseEventHandler, useEffect, useState } from 'react';
import { FormProvider, useForm, useFormContext } from 'react-hook-form';
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
  formatService,
  Integration,
  IntegrationConfig,
  Service,
  SupportedIntegrations,
} from '../../../utils/integrations';
import { isFailed, isLoading, isSucceeded } from '../../../utils/shared';
import { IntegrationTextInputField } from './IntegrationTextInputField';

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
  const [showDialog, setShowDialog] = useState<boolean>(true);
  const [submitDisabled, setSubmitDisabled] = useState<boolean>(true);
  const editMode = !!integrationToEdit;
  const dispatch: AppDispatch = useDispatch();

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

  // Make sure that the user object is ready.
  if (!user) {
    return null;
  }

  const numWorkflows = new Set(operators.map((x) => x.workflow_id)).size;

  const connectStatus = editMode ? editStatus : connectNewStatus;

  const onConfirmDialog = (
    data: IntegrationConfig,
    user: UserProfile,
    editMode = false,
    integrationId?: string
  ) => {
    if (!editMode) {
      for (let i = 0; i < integrations.length; i++) {
        if (data.name === integrations[i].name) {
          setShouldShowNameError(true);
          return;
        }
      }
    }

    // We do this so we can collect name form inputs inside the same form context.
    const name = data.name;
    // remove the name key from data so pydantic doesn't throw error.
    delete data.name;

    return editMode
      ? dispatch(
          handleEditIntegration({
            apiKey: user.apiKey,
            integrationId: integrationId,
            name: name,
            config: data,
          })
        )
      : dispatch(
          handleConnectToNewIntegration({
            apiKey: user.apiKey,
            service: service,
            name: name,
            config: data,
          })
        );
  };

  const [migrateStorage, setMigrateStorage] = useState(false);

  const methods = useForm({
    resolver: yupResolver(validationSchema),
  });

  // Check to enable/disable submit button
  useEffect(() => {
    const subscription = methods.watch(async () => {
      const checkIsFormValid = async () => {
        const isValidForm = await methods.trigger();
        if (isValidForm && submitDisabled) {
          // Form is valid, enable the submit button.
          setSubmitDisabled(false);
        } else {
          // Form is still invalid, disable the submit button.
          setSubmitDisabled(true);
        }
      };

      checkIsFormValid();
    });

    // Unsubscribe and handle lifecycle changes.
    return () => subscription.unsubscribe();
  }, [methods.watch]);

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

  const handleCloseDialog = () => {
    if (onCloseDialog) {
      onCloseDialog();
    }
    setShowDialog(false);
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
    <Dialog
      open={showDialog}
      onClose={handleCloseDialog}
      fullWidth
      maxWidth="lg"
    >
      <FormProvider {...methods}>
        <form>
          {service !== 'Kubernetes' && (
            <DialogHeader
              integrationToEdit={integrationToEdit}
              service={service}
            />
          )}
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

            {service !== 'Kubernetes' && nameInput}

            {dialogContent({
              user,
              editMode,
              onCloseDialog: handleCloseDialog,
              loading: isLoading(connectStatus),
              disabled: submitDisabled,
            })}

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
          {service !== 'Kubernetes' && (
            <DialogActionButtons
              onCloseDialog={handleCloseDialog}
              loading={isLoading(connectStatus)}
              disabled={submitDisabled}
              onSubmit={async () => {
                await methods.handleSubmit((data, event) => {
                  return onConfirmDialog(
                    data,
                    user,
                    editMode,
                    integrationToEdit?.id
                  );
                })(); // Remember the last two parens to call the function!
              }}
            />
          )}
        </form>
      </FormProvider>
      <DevTool control={methods.control} />{' '}
      {/* set up the dev tool for debugging forms (only runs in dev mode) */}
    </Dialog>
  );
};

type DialogActionButtonProps = {
  onCloseDialog: () => void;
  loading: boolean;
  disabled: boolean;
  onSubmit: MouseEventHandler<HTMLButtonElement> | undefined;
};

export const DialogActionButtons: React.FC<DialogActionButtonProps> = ({
  user,
  editMode = false,
  onCloseDialog,
  loading,
  disabled,
  onSubmit,
}) => {
  const methods = useFormContext();
  return (
    <DialogActions>
      <Button autoFocus onClick={onCloseDialog}>
        Cancel
      </Button>
      <LoadingButton
        autoFocus
        onClick={() => {
          onSubmit();
        }}
        loading={loading}
        disabled={disabled}
      >
        Confirm
      </LoadingButton>
    </DialogActions>
  );
};

const getConnectionMessage = (service: Service) => {
  if (service === 'AWS') {
    return 'Configuring Aqueduct-managed Kubernetes on AWS';
  } else {
    return `Connecting to ${service}`;
  }
};

type DialogHeaderProps = {
  integrationToEdit: Integration | undefined;
  service: Service;
};
export const DialogHeader: React.FC<DialogHeaderProps> = ({
  integrationToEdit,
  service,
}) => {
  const connectionMessage = getConnectionMessage(service);

  return (
    <DialogTitle>
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
    </DialogTitle>
  );
};

export default IntegrationDialog;
