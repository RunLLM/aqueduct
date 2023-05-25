import { faFileCsv } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
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
import React, { useState } from 'react';

import UserProfile from '../../../utils/auth';
import { addTable, CSVConfig } from '../../../utils/resources';
import { CSVDialog } from './csvDialog';

type Props = {
  user: UserProfile;
  resourceId: string;
  onCloseDialog: () => void;
  onConnect: () => void;
};

const AddTableDialog: React.FC<Props> = ({
  user,
  resourceId,
  onCloseDialog,
  onConnect,
}) => {
  const [config, setConfig] = useState<CSVConfig>({
    name: '',
    csv: null,
  });
  const [successMessage, setSuccessMessage] = useState('');
  const [showSuccessToast, setShowSuccessToast] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);
  const [errMsg, setErrMsg] = useState(null);

  const handleSuccessToastClose = () => {
    setShowSuccessToast(false);
  };

  const disabled = !config.name || !config.csv;

  const dialogHeader = (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'row',
        justifyContent: 'space-between',
        width: '100%',
      }}
    >
      <Typography variant="h5">{'Upload CSV'}</Typography>
      <FontAwesomeIcon icon={faFileCsv} size="2x" color="black" />
    </Box>
  );

  const serviceDialog = (
    <CSVDialog setDialogConfig={setConfig} setErrMsg={setErrMsg} />
  );

  const confirmConnect = () => {
    setIsConnecting(true);
    setErrMsg(null);

    addTable(user, resourceId, config)
      .then(() => {
        setShowSuccessToast(true);
        const successMessage =
          'Successfully uploaded CSV file to the demo database!';
        setSuccessMessage(successMessage);
        onConnect();
        setIsConnecting(false);
      })
      .catch((err) => {
        setErrMsg(err.message);
        setIsConnecting(false);
      });
  };

  return (
    <Dialog open={true} onClose={onCloseDialog} fullWidth maxWidth="lg">
      <DialogTitle>{dialogHeader}</DialogTitle>
      <DialogContent>
        {serviceDialog}
        {errMsg && (
          <Alert severity="error">
            <AlertTitle>Unable to upload CSV file to demo database.</AlertTitle>
            <pre>{errMsg}</pre>
          </Alert>
        )}
        <Snackbar
          anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
          open={showSuccessToast}
          onClose={handleSuccessToastClose}
          key={'resources-dialog-success-snackbar'}
          autoHideDuration={6000}
        >
          <Alert
            onClose={handleSuccessToastClose}
            severity="success"
            sx={{ width: '100%' }}
          >
            {successMessage}
          </Alert>
        </Snackbar>
      </DialogContent>
      <DialogActions>
        <Button autoFocus onClick={onCloseDialog}>
          Cancel
        </Button>
        <LoadingButton
          autoFocus
          onClick={confirmConnect}
          loading={isConnecting}
          disabled={disabled}
        >
          Confirm
        </LoadingButton>
      </DialogActions>
    </Dialog>
  );
};

export default AddTableDialog;
