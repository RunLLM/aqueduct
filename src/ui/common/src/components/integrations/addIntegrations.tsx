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
  const dispatch: AppDispatch = useDispatch();
  const [showSuccessToast, setShowSuccessToast] = useState<Service>(null);
  const handleSuccessToastClose = () => {
    setShowSuccessToast(null);
  };

  return (
    <Box sx={{ maxWidth: '616px' }}>
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
            const [showDialog, setShowDialog] = useState(false);

            const iconWrapper = (
              <Box
                onClick={() => setShowDialog(integration.activated)}
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
                <img
                  src={integration.logo}
                  width="100%"
                  style={{
                    opacity: integration.activated ? 1.0 : 0.3,
                    height: '85px',
                    width: '160px',
                    maxWidth: '160px',
                    maxHeight: '85px',
                    objectFit: 'contain',
                  }}
                />
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
