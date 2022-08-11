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
import React, { useEffect, useState } from 'react';

import UserProfile from '../../../utils/auth';
import { addTable, CSVConfig } from '../../../utils/integrations';
import { CSVDialog } from './csvDialog';
import { isConfigComplete } from './dialog';

type Props = {
  user: UserProfile;
  integrationId: string;
  onCloseDialog: () => void;
  onConnect: () => void;
};

const AddTableDialog: React.FC<Props> = ({
  user,
  integrationId,
  onCloseDialog,
  onConnect,
}) => {
  const [config, setConfig] = useState<CSVConfig>({
    name: '',
    csv: null,
  });
  const [disableConnect, setDisableConnect] = useState(true);
  const [successMessage, setSuccessMessage] = useState('');
  const [showSuccessToast, setShowSuccessToast] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);
  const [errMsg, setErrMsg] = useState(null);

  const handleSuccessToastClose = () => {
    setShowSuccessToast(false);
  };

  useEffect(() => {
    setDisableConnect(!isConfigComplete(config));
  }, [config]);

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

    addTable(user, integrationId, config)
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
          key={'integrations-dialog-success-snackbar'}
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
          disabled={disableConnect}
        >
          Confirm
        </LoadingButton>
      </DialogActions>
    </Dialog>
  );
};

export default AddTableDialog;
