import { Alert, Snackbar } from '@mui/material';
import React, { useEffect, useState } from 'react';

type ErrorSnackbarProps = {
  shouldShow: boolean;
  errMsg: string;
};

export const ErrorSnackbar: React.FC<ErrorSnackbarProps> = ({
  shouldShow,
  errMsg,
}) => {
  const [showErrorToast, setShowErrorToast] = useState(false);
  useEffect(() => {
    if (shouldShow) {
      setShowErrorToast(true);
    } else {
      // Remember to hide this if the error goes away.
      setShowErrorToast(false);
    }
  }, [shouldShow]);

  const handleErrorToastClose = () => {
    setShowErrorToast(false);
  };

  return (
    <Snackbar
      anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      open={showErrorToast}
      onClose={handleErrorToastClose}
      key={'failure-snackbar'}
      autoHideDuration={6000}
    >
      <Alert
        onClose={handleErrorToastClose}
        severity="error"
        sx={{ width: '100%' }}
      >
        {errMsg}
      </Alert>
    </Snackbar>
  );
};
