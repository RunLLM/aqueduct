import { LoadingButton } from '@mui/lab';
import { Alert, DialogActions, DialogContent } from '@mui/material';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import { isEqual } from 'lodash';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';

import { useAqueductConsts } from '../../../components/hooks/useAqueductConsts';
import {
  handleDeleteIntegration,
  resetDeletionStatus,
} from '../../../reducers/integration';
import { AppDispatch, RootState } from '../../../stores/store';
import UserProfile from '../../../utils/auth';
import { IntegrationConfig, Service } from '../../../utils/integrations';
import { isFailed, isLoading, isSucceeded } from '../../../utils/shared';
import {
  convertIntegrationConfigToMetadataStorageConfig,
  getMetadataStorageConfig,
  MetadataStorageConfig,
} from '../../../utils/storage';

type Props = {
  user: UserProfile;
  integrationId: string;
  integrationName: string;
  integrationType: Service;
  config: IntegrationConfig;
  onCloseDialog: () => void;
};

const DeleteIntegrationDialog: React.FC<Props> = ({
  user,
  integrationId,
  integrationName,
  integrationType,
  config,
  onCloseDialog,
}) => {
  const dispatch: AppDispatch = useDispatch();
  const navigate = useNavigate();
  const [isConnecting, setIsConnecting] = useState(false);

  const { apiAddress } = useAqueductConsts();

  const [metadataStorageConfig, setMetadataStorageConfig] =
    useState<MetadataStorageConfig | null>(null);

  useEffect(() => {
    async function fetchMetadataStorageConfig() {
      const metadataStorageConfig = await getMetadataStorageConfig(
        apiAddress,
        user.apiKey
      );
      setMetadataStorageConfig(metadataStorageConfig);
    }

    fetchMetadataStorageConfig();
  }, []);

  const deleteIntegrationStatus = useSelector(
    (state: RootState) => state.integrationReducer.deletionStatus
  );

  useEffect(() => {
    if (!isLoading(deleteIntegrationStatus)) {
      setIsConnecting(false);
    }

    if (isSucceeded(deleteIntegrationStatus)) {
      navigate('/integrations', {
        state: {
          deleteIntegrationStatus: deleteIntegrationStatus,
          deleteIntegrationName: integrationName,
        },
      });
    }
  }, [deleteIntegrationStatus, integrationName, navigate]);

  const confirmConnect = () => {
    setIsConnecting(true);
    dispatch(
      handleDeleteIntegration({
        apiKey: user.apiKey,
        integrationId: integrationId,
      })
    );
  };

  const operatorsState = useSelector((state: RootState) => {
    return state.integrationReducer.operators;
  });

  const isStorage = config.use_as_storage;
  let isCurrentStorage = true;
  if (isStorage && metadataStorageConfig) {
    const storageConfig = convertIntegrationConfigToMetadataStorageConfig(
      config,
      metadataStorageConfig,
      integrationType
    );
    // Check deep equality
    isCurrentStorage = isEqual(storageConfig, metadataStorageConfig);
  }

  if (isCurrentStorage) {
    return (
      <Dialog
        open={!deleteIntegrationStatus || !isFailed(deleteIntegrationStatus)}
        onClose={onCloseDialog}
        maxWidth="lg"
      >
        <DialogContent>
          We cannot delete this integration because it is acting as the metadata
          storage location.
        </DialogContent>
        <DialogActions>
          <Button onClick={onCloseDialog}>Dismiss</Button>
        </DialogActions>
      </Dialog>
    );
  } else if (
    isSucceeded(operatorsState.status) &&
    !operatorsState.operators.some((op) => op.is_active)
  ) {
    return (
      <>
        <Dialog
          open={!deleteIntegrationStatus || !isFailed(deleteIntegrationStatus)}
          onClose={onCloseDialog}
          maxWidth="lg"
        >
          <DialogContent>
            Are you sure you want to delete the integration?
          </DialogContent>
          <DialogActions>
            <Button onClick={onCloseDialog}>Cancel</Button>
            <LoadingButton
              autoFocus
              onClick={confirmConnect}
              loading={isConnecting}
            >
              Confirm
            </LoadingButton>
          </DialogActions>
        </Dialog>
        <Dialog
          open={isFailed(deleteIntegrationStatus)}
          onClose={onCloseDialog}
          maxWidth="lg"
        >
          {deleteIntegrationStatus && isFailed(deleteIntegrationStatus) && (
            <Alert severity="error" sx={{ margin: 2 }}>
              Integration deletion failed with error:
              <br></br>
              <pre>{deleteIntegrationStatus.err}</pre>
            </Alert>
          )}
          <DialogActions>
            <Button
              onClick={() => {
                onCloseDialog();
                dispatch(resetDeletionStatus());
              }}
            >
              Dismiss
            </Button>
          </DialogActions>
        </Dialog>
      </>
    );
  } else {
    return (
      <Dialog
        open={!deleteIntegrationStatus || !isFailed(deleteIntegrationStatus)}
        onClose={onCloseDialog}
        maxWidth="lg"
      >
        <DialogContent>
          We cannot delete this integration because it is currently being used
          by at least one workflow.
        </DialogContent>
        <DialogActions>
          <Button onClick={onCloseDialog}>Dismiss</Button>
        </DialogActions>
      </Dialog>
    );
  }
};

export default DeleteIntegrationDialog;
