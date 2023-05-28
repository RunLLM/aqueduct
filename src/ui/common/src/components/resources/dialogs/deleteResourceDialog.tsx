import { LoadingButton } from '@mui/lab';
import { Alert, DialogActions, DialogContent } from '@mui/material';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';

import { handleGetServerConfig } from '../../../handlers/getServerConfig';
import {
  handleDeleteResource,
  handleLoadResourceOperators,
  resetDeletionStatus,
} from '../../../reducers/resource';
import { AppDispatch, RootState } from '../../../stores/store';
import UserProfile from '../../../utils/auth';
import {
  AqueductComputeConfig,
  aqueductComputeName,
  ResourceConfig,
  Service,
} from '../../../utils/resources';
import { isFailed, isLoading, isSucceeded } from '../../../utils/shared';
import { convertResourceConfigToServerConfig } from '../../../utils/storage';

const isEqual = function (x, y) {
  if (x === y) {
    return true;
  } else if (
    typeof x == 'object' &&
    x != null &&
    typeof y == 'object' &&
    y != null
  ) {
    if (Object.keys(x).length != Object.keys(y).length) {
      return false;
    }

    for (const prop in x) {
      if (y.hasOwnProperty(prop)) {
        if (!isEqual(x[prop], y[prop])) {
          return false;
        }
      } else {
        return false;
      }
    }
    return true;
  } else {
    return false;
  }
};

type Props = {
  user: UserProfile;
  resourceId: string;
  resourceName: string;
  resourceType: Service;
  config: ResourceConfig;
  onCloseDialog: () => void;
};

const DeleteResourceDialog: React.FC<Props> = ({
  user,
  resourceId,
  resourceName,
  resourceType,
  config,
  onCloseDialog,
}) => {
  // If the resource is the Aqueduct Server, we need to translate the fields so that
  // we delete the registered Conda resource, not the Aqueduct Server itself. Deleting
  // the vanilla Aqueduct Server is not possible.
  if (resourceName === aqueductComputeName) {
    const aqConfig = config as AqueductComputeConfig;
    resourceId = aqConfig.conda_resource_id;
    resourceName = aqConfig.conda_resource_name;
    resourceType = 'Conda';
    config = JSON.parse(aqConfig.conda_config_serialized);
  }

  const dispatch: AppDispatch = useDispatch();
  const navigate = useNavigate();
  const [isConnecting, setIsConnecting] = useState(false);

  const serverConfig = useSelector(
    (state: RootState) => state.serverConfigReducer
  );

  useEffect(() => {
    async function fetchServerConfig() {
      await dispatch(handleGetServerConfig({ apiKey: user.apiKey }));
    }

    async function fetchLoadResourceOperators() {
      await dispatch(
        handleLoadResourceOperators({
          apiKey: user.apiKey,
          resourceId: resourceId,
        })
      );
    }

    fetchServerConfig();
    fetchLoadResourceOperators();
  }, []);

  const deleteResourceStatus = useSelector(
    (state: RootState) => state.resourceReducer.deletionStatus
  );

  useEffect(() => {
    if (!isLoading(deleteResourceStatus)) {
      setIsConnecting(false);
    }

    if (isSucceeded(deleteResourceStatus)) {
      navigate('/resources', {
        state: {
          deleteResourceStatus: deleteResourceStatus,
          deleteResourceName: resourceName,
        },
      });
    }
  }, [deleteResourceStatus, resourceName, navigate]);

  const confirmConnect = () => {
    setIsConnecting(true);
    dispatch(
      handleDeleteResource({
        apiKey: user.apiKey,
        resourceId: resourceId,
      })
    );
  };

  const operatorsState = useSelector((state: RootState) => {
    return state.resourceReducer.operators;
  });

  const isStorage = config.use_as_storage === 'true';
  let isCurrentStorage = isStorage;
  if (isStorage && serverConfig) {
    const storageConfig = convertResourceConfigToServerConfig(
      config,
      serverConfig.config,
      resourceType
    );
    // Check deep equality
    isCurrentStorage = isEqual(storageConfig, serverConfig.config);
  }
  if (isCurrentStorage) {
    return (
      <Dialog
        open={!deleteResourceStatus || !isFailed(deleteResourceStatus)}
        onClose={onCloseDialog}
        maxWidth="lg"
      >
        <DialogContent>
          We cannot delete this resource because it is acting as the metadata
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
          open={!deleteResourceStatus || !isFailed(deleteResourceStatus)}
          onClose={onCloseDialog}
          maxWidth="lg"
        >
          <DialogContent>
            Are you sure you want to delete the resource?
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
          open={isFailed(deleteResourceStatus)}
          onClose={onCloseDialog}
          maxWidth="lg"
        >
          {deleteResourceStatus && isFailed(deleteResourceStatus) && (
            <Alert severity="error" sx={{ margin: 2 }}>
              Resource deletion failed with error:
              <br></br>
              <pre>{deleteResourceStatus.err}</pre>
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
        open={!deleteResourceStatus || isFailed(deleteResourceStatus)}
        onClose={onCloseDialog}
        maxWidth="lg"
      >
        <DialogContent>
          We cannot delete this resource because it is currently being used by
          at least one workflow.
        </DialogContent>
        <DialogActions>
          <Button onClick={onCloseDialog}>Dismiss</Button>
        </DialogActions>
      </Dialog>
    );
  }
};

export default DeleteResourceDialog;
