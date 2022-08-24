import { LoadingButton } from '@mui/lab';
import { Alert, DialogActions, DialogContent } from '@mui/material';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';

import {
  handleDeleteIntegration,
  resetDeletionStatus,
} from '../../../reducers/integration';
import { AppDispatch, RootState } from '../../../stores/store';
import UserProfile from '../../../utils/auth';
import { isFailed, isLoading, isSucceeded } from '../../../utils/shared';
type Props = {
  user: UserProfile;
  integrationId: string;
  integrationName: string;
  onCloseDialog: () => void;
};

const DeleteIntegrationDialog: React.FC<Props> = ({
  user,
  integrationId,
  integrationName,
  onCloseDialog,
}) => {
  const dispatch: AppDispatch = useDispatch();
  const navigate = useNavigate();

  const [isConnecting, setIsConnecting] = useState(false);

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
  }, [deleteIntegrationStatus]);

  const confirmConnect = () => {
    setIsConnecting(true);
    dispatch(
      handleDeleteIntegration({
        apiKey: user.apiKey,
        integrationId: integrationId,
      })
    );
  };

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
};

export default DeleteIntegrationDialog;
