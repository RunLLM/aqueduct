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
  handleConnectToNewResource,
  handleEditResource,
} from '../../../reducers/resource';
import { handleLoadResources } from '../../../reducers/resources';
import { AppDispatch, RootState } from '../../../stores/store';
import UserProfile from '../../../utils/auth';
import {
  formatService,
  Resource,
  ResourceConfig,
  Service,
} from '../../../utils/resources';
import { isFailed, isLoading, isSucceeded } from '../../../utils/shared';
import SupportedResources from '../../../utils/SupportedResources';
import { ResourceTextInputField } from './ResourceTextInputField';

type Props = {
  user: UserProfile;
  service: Service;
  onCloseDialog: () => void;
  onSuccess: () => void;
  showMigrationDialog?: () => void;
  resourceToEdit?: Resource;
  dialogContent: React.FC;
  validationSchema: Yup.ObjectSchema<any>;
};

const ResourceDialog: React.FC<Props> = ({
  user,
  service,
  onCloseDialog,
  onSuccess,
  showMigrationDialog = undefined,
  resourceToEdit = undefined,
  dialogContent,
  validationSchema,
}) => {
  const [showDialog, setShowDialog] = useState<boolean>(true);
  const [submitDisabled, setSubmitDisabled] = useState<boolean>(true);
  const [migrateStorage, setMigrateStorage] = useState(false);

  const methods = useForm({
    resolver: yupResolver(validationSchema),
  });

  const editMode = !!resourceToEdit;
  const dispatch: AppDispatch = useDispatch();

  const [shouldShowNameError, setShouldShowNameError] =
    useState<boolean>(false);

  const connectNewStatus = useSelector(
    (state: RootState) => state.resourceReducer.connectNewStatus
  );

  const editStatus = useSelector(
    (state: RootState) => state.resourceReducer.editStatus
  );

  const operators = useSelector(
    (state: RootState) => state.resourceReducer.operators.operators
  );

  const resources = useSelector((state: RootState) =>
    Object.values(state.resourcesReducer.resources)
  );

  const numWorkflows = new Set(operators.map((x) => x.workflow_id)).size;

  const connectStatus = editMode ? editStatus : connectNewStatus;

  const onConfirmDialog = (
    data: ResourceConfig,
    user: UserProfile,
    editMode = false,
    resourceId?: string
  ) => {
    if (!editMode) {
      for (let i = 0; i < resources.length; i++) {
        if (data.name === resources[i].name) {
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
          handleEditResource({
            apiKey: user.apiKey,
            resourceId: resourceId,
            name: name,
            config: data,
          })
        )
      : dispatch(
          handleConnectToNewResource({
            apiKey: user.apiKey,
            service: service,
            name: name,
            config: data,
          })
        );
  };

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
        handleLoadResources({ apiKey: user.apiKey, forceLoad: true })
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
    <ResourceTextInputField
      name="name"
      spellCheck={false}
      required={true}
      label="Name*"
      description="Provide a unique name to refer to this resource."
      placeholder={'my_' + formatService(service) + '_resource'}
      onChange={(event) => {
        setShouldShowNameError(false);
        methods.setValue('name', event.target.value);
      }}
    />
  );

  // Make sure that the user object is ready.
  if (!user) {
    return null;
  }

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
              resourceToEdit={resourceToEdit}
              service={service}
            />
          )}
          <DialogContent>
            {editMode && numWorkflows > 0 && (
              <Alert sx={{ mb: 2 }} severity="info">
                {`Changing this resource will automatically update ${numWorkflows} ${
                  numWorkflows === 1 ? 'workflow' : 'workflows'
                }.`}
              </Alert>
            )}
            {(service === 'Email' || service === 'Slack') && (
              <Typography variant="body1" color="gray.700">
                To learn more about how to set up {service}, see our{' '}
                <Link
                  href={SupportedResources[service].docs}
                  target="_blank"
                >
                  documentation
                </Link>
                .
              </Typography>
            )}

            {service !== 'Kubernetes' && service !== 'Conda' && nameInput}

            {dialogContent({
              user,
              editMode,
              onCloseDialog: handleCloseDialog,
              loading: isLoading(connectStatus),
              disabled: submitDisabled,
              setMigrateStorage,
            })}

            {shouldShowNameError && (
              <Alert sx={{ mt: 2 }} severity="error">
                <AlertTitle>Naming Error</AlertTitle>A connected resource
                already exists with this name. Please provide a unique name for
                your resource.
              </Alert>
            )}

            {isFailed(connectStatus) && (
              <Alert sx={{ mt: 2 }} severity="error">
                <AlertTitle>
                  {editMode
                    ? `Failed to update ${resourceToEdit.name}`
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
                    resourceToEdit?.id
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
  resourceToEdit: Resource | undefined;
  service: Service;
};
export const DialogHeader: React.FC<DialogHeaderProps> = ({
  resourceToEdit,
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
          {!!resourceToEdit
            ? `Edit ${resourceToEdit.name}`
            : `${connectionMessage}`}
        </Typography>
        <img height="45px" src={SupportedResources[service].logo} />
      </Box>
    </DialogTitle>
  );
};

export default ResourceDialog;
