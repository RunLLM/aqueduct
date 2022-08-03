import { Alert, DialogActions, DialogContent, Typography } from '@mui/material';
import CircularProgress from '@mui/material/CircularProgress';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import React from 'react';
import { useSelector } from 'react-redux';

import { RootState } from '../../../stores/store';
import {
  GithubIssueLink,
  isFailed,
  isLoading,
  isSucceeded,
} from '../../../utils/shared';
import { Button } from '../../primitives/Button.styles';

type Props = {
  onCloseDialog: () => void;
};

const TestConnectDialog: React.FC<Props> = ({ onCloseDialog }) => {
  const status = useSelector(
    (state: RootState) => state.integrationReducer.connectionStatus
  );

  let dialogBody = null;
  if (isLoading(status)) {
    dialogBody = <CircularProgress />;
  }

  if (isSucceeded(status)) {
    dialogBody = (
      <Alert severity="success">Test-connection is successful!</Alert>
    );
  }

  if (isFailed(status)) {
    dialogBody = (
      <Alert severity="error">
        {`Test-connect failed with error: ${status.err}\n Please confirm that your integration is running.`}
      </Alert>
    );
  }

  return (
    <Dialog open={true} onClose={onCloseDialog} fullWidth maxWidth="lg">
      <DialogTitle>
        <Typography variant="h5">{'Testing Connection'}</Typography>
      </DialogTitle>
      <DialogContent>{dialogBody}</DialogContent>
      <DialogActions>
        {isLoading(status) && (
          <Button color="secondary" autoFocus onClick={onCloseDialog}>
            Cancel
          </Button>
        )}
        {isFailed(status) && (
          <Button
            autoFocus
            color="secondary"
            href={GithubIssueLink}
            onClick={onCloseDialog}
          >
            {'File an issue'}
          </Button>
        )}
        {(isSucceeded(status) || isFailed(status)) && (
          <Button
            color={isSucceeded ? 'success' : 'primary'}
            autoFocus
            onClick={onCloseDialog}
          >
            {'Confirm'}
          </Button>
        )}
      </DialogActions>
    </Dialog>
  );
};

export default TestConnectDialog;
