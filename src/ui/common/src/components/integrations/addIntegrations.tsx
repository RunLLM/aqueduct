import { Typography } from '@mui/material';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import Snackbar from '@mui/material/Snackbar';
import React, { useState } from 'react';
import { useDispatch } from 'react-redux';

import { resetConnectNewStatus } from '../../reducers/integration';
import { AppDispatch } from '../../stores/store';
import { theme } from '../../styles/theme/theme';
import UserProfile from '../../utils/auth';
import { Service, ServiceInfoMap } from '../../utils/integrations';
import IntegrationDialog from './dialogs/dialog';
import IntegrationLogo from './logo';

type Props = {
  user: UserProfile;
  supportedIntegrations: ServiceInfoMap;
  category: string;
};

const AddIntegrations: React.FC<Props> = ({
  user,
  supportedIntegrations,
  category,
}) => {
  const [showDialog, setShowDialog] = useState(false);
  const dispatch: AppDispatch = useDispatch();
  const [showSuccessToast, setShowSuccessToast] = useState<Service>(null);
  const handleSuccessToastClose = () => {
    setShowSuccessToast(null);
  };
  const [showMigrationDialog, setShowMigrationDialog] = useState(false);

  return (
    <Box sx={{ maxWidth: '616px' }}>
      {showMigrationDialog && (
        <Alert
          onClose={() => {
            setShowMigrationDialog(false);
          }}
          severity="info"
          sx={{ width: '100%' }}
        >
          {`Storage migration is in progress. The server will be temporarily unavailable. Please refresh the page to check if the server is ready.`}
        </Alert>
      )}
      <Grid
        container
        spacing={1}
        sx={{ my: '16px', width: '100%' }}
        columns={4}
      >
        {Object.entries(supportedIntegrations)
          .filter(([svc]) => svc !== 'Aqueduct Demo')
          .map(([svc, integration]) => {
            if (integration.category !== category) {
              return null;
            }

            const service = svc as Service;
            const iconWrapper = (
              <Box
                onClick={() => {
                  setShowDialog(integration.activated);
                }}
                sx={{
                  width: '160px',
                  height: '128px',
                  px: 2,
                  py: 2,
                  borderRadius: 2,
                  border: `2px solid ${theme.palette.gray['700']}`,
                  cursor: integration.activated ? 'pointer' : 'default',
                  '&:hover': {
                    backgroundColor: integration.activated
                      ? 'gray.300'
                      : 'white',
                  },
                }}
              >
                <Box
                  width="160px"
                  maxWidth="160px"
                  display="flex"
                  flexDirection="column"
                  alignItems="center"
                >
                  <IntegrationLogo
                    service={service}
                    activated={integration.activated}
                    size="large"
                  />
                </Box>
                <Typography
                  variant={'body1'}
                  align={'center'}
                  sx={{ marginTop: '16px' }}
                >
                  {service}
                </Typography>
              </Box>
            );

            return (
              <Grid container item xs={4} key={service}>
                <Box>
                  {iconWrapper}
                  {showDialog && (
                    <IntegrationDialog
                      user={user}
                      service={service}
                      onSuccess={() => {
                        setShowSuccessToast(service);
                      }}
                      onCloseDialog={() => {
                        setShowDialog(false);
                        dispatch(resetConnectNewStatus());
                      }}
                      showMigrationDialog={() => setShowMigrationDialog(true)}
                    />
                  )}
                </Box>
                <Snackbar
                  anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
                  open={showSuccessToast === service}
                  onClose={handleSuccessToastClose}
                  key={'integrations-dialog-success-snackbar'}
                  autoHideDuration={6000}
                >
                  <Alert
                    onClose={handleSuccessToastClose}
                    severity="success"
                    sx={{ width: '100%' }}
                  >
                    {`Successfully connected to ${service}!`}
                  </Alert>
                </Snackbar>
              </Grid>
            );
          })}
      </Grid>
    </Box>
  );
};

export default AddIntegrations;
