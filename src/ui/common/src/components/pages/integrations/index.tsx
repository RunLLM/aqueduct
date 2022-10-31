import { Alert, Snackbar, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import Divider from '@mui/material/Divider';
import React, { useEffect, useState } from 'react';
import { useLocation } from 'react-router-dom';

import { BreadcrumbLink } from '../../../components/layouts/NavBar';
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
  let forceLoad = false;

  const [
    showDeleteIntegrationSuccessToast,
    setShowDeleteIntegrationSuccessToast,
  ] = useState(false);

  if (location.state && location.state !== undefined) {
    const navState = location.state as integrationsNavigateState;
    deleteIntegrationName = navState.deleteIntegrationName;
    if (!showDeleteIntegrationSuccessToast) {
      setShowDeleteIntegrationSuccessToast(
        navState.deleteIntegrationStatus.loading === LoadingStatusEnum.Succeeded
      );
    }

    // Reload integrations because deleted
    forceLoad = true;
  }

  return (
    <Layout
      breadcrumbs={[BreadcrumbLink.HOME, BreadcrumbLink.INTEGRATIONS]}
      user={user}
    >
      <Box>
        <Typography variant="h2" gutterBottom component="div">
          Integrations
        </Typography>

        <Divider sx={{ width: '950px' }} />

        <Box sx={{ my: 3 }}>
          <Typography variant="h4" marginBottom={3}>
            Add an Integration
          </Typography>
          <Typography variant="h6" marginBottom={2}>
            Data
          </Typography>
          <AddIntegrations
            user={user}
            category="data"
            supportedIntegrations={SupportedIntegrations}
          />
          <Typography variant="h6" marginTop={4} marginBottom={3}>
            Compute
          </Typography>
          <AddIntegrations
            user={user}
            category="compute"
            supportedIntegrations={SupportedIntegrations}
          />
        </Box>

        <Box marginTop="40px">
          <Divider sx={{ width: '950px' }} />
        </Box>

        <Box sx={{ marginTop: 3, marginBottom: 3 }}>
          <Typography variant="h4" marginBottom={2}>
            Connected Integrations
          </Typography>
          <ConnectedIntegrations user={user} forceLoad={forceLoad} />
        </Box>
      </Box>

      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={showDeleteIntegrationSuccessToast}
        key={'workflowheader-delete-success-error-snackbar'}
        autoHideDuration={6000}
        onClose={() => {
          setShowDeleteIntegrationSuccessToast(false);
          location.state = undefined;
        }}
      >
        <Alert severity="success" sx={{ width: '100%' }}>
          {`Successfully deleted ${deleteIntegrationName}`}
        </Alert>
      </Snackbar>
    </Layout>
  );
};

export default IntegrationsPage;
