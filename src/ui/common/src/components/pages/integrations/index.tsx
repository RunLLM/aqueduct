import { Alert, Snackbar, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import Divider from '@mui/material/Divider';
import React, { useEffect } from 'react';
import { useLocation } from 'react-router-dom';

import UserProfile from '../../../utils/auth';
import { SupportedIntegrations } from '../../../utils/integrations';
import { LoadingStatus, LoadingStatusEnum } from '../../../utils/shared';
import AddIntegrations from '../../integrations/addIntegrations';
import { ConnectedIntegrations } from '../../integrations/connectedIntegrations';
import DefaultLayout from '../../layouts/default';
import { LayoutProps } from '../types';

type Props = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

type integrationsNavigateState = {
  deleteIntegrationStatus: LoadingStatus;
  deleteIntegrationName: string;
};

const IntegrationsPage: React.FC<Props> = ({
  user,
  Layout = DefaultLayout,
}) => {
  const location = useLocation();

  useEffect(() => {
    document.title = 'Integrations | Aqueduct';
  }, []);

  let deleteIntegrationName = '';
  let openDeleteIntegrationSuccessSnackbar = false;
  let forceLoad = false;

  if (location.state && location.state !== undefined) {
    const navState = location.state as integrationsNavigateState;
    deleteIntegrationName = navState.deleteIntegrationName;
    openDeleteIntegrationSuccessSnackbar =
      navState.deleteIntegrationStatus.loading === LoadingStatusEnum.Succeeded;
    // Reload integrations because deleted
    forceLoad = true;
  }

  return (
    <Layout user={user}>
      <Box>
        {/*<Breadcrumbs>
          <Link
            underline="hover"
            color="inherit"
            to="/"
            component={RouterLink as any}
          >
            Home
          </Link>
          <Typography color="text.primary">Integrations</Typography>
        </Breadcrumbs>*/}

        <Typography variant="h2" gutterBottom component="div">
          Integrations
        </Typography>

        <Box sx={{ my: 3, ml: 1 }}>
          <Typography variant="h4">Add an Integration</Typography>
          <Typography variant="h6">Data</Typography>
          <AddIntegrations
            user={user}
            category="data"
            supportedIntegrations={SupportedIntegrations}
          />
          <Typography variant="h6">Compute</Typography>
          <AddIntegrations
            user={user}
            category="compute"
            supportedIntegrations={SupportedIntegrations}
          />
        </Box>

        <Divider sx={{ width: '950px' }} />

        <Box sx={{ my: 3, ml: 1 }}>
          <Typography variant="h4">Connected Integrations</Typography>
          <ConnectedIntegrations user={user} forceLoad={forceLoad} />
        </Box>
      </Box>
      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={openDeleteIntegrationSuccessSnackbar}
        key={'workflowheader-delete-success-error-snackbar'}
        autoHideDuration={6000}
      >
        <Alert severity="success" sx={{ width: '100%' }}>
          {`Successfully deleted ${deleteIntegrationName}`}
        </Alert>
      </Snackbar>
    </Layout>
  );
};

export default IntegrationsPage;
