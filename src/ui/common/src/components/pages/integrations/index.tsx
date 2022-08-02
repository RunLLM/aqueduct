import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import Breadcrumbs from '@mui/material/Breadcrumbs';
import Divider from '@mui/material/Divider';
import Link from '@mui/material/Link';
import React, { useEffect } from 'react';
import { Link as RouterLink } from 'react-router-dom';

import UserProfile from '../../../utils/auth';
import { SupportedIntegrations } from '../../../utils/integrations';
import AddIntegrations from '../../integrations/addIntegrations';
import { ConnectedIntegrations } from '../../integrations/connectedIntegrations';
import DefaultLayout from '../../layouts/default';
import { LayoutProps } from '../types';

type Props = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

const IntegrationsPage: React.FC<Props> = ({
  user,
  Layout = DefaultLayout,
}) => {
  useEffect(() => {
    document.title = 'Integrations | Aqueduct';
  }, []);

  return (
    <Layout user={user}>
      <Box>
        <Breadcrumbs>
          <Link
            underline="hover"
            color="inherit"
            to="/"
            component={RouterLink as any}
          >
            Home
          </Link>
          <Link
            underline="hover"
            color="inherit"
            to="/integrations"
            component={RouterLink as any}
          >
            Integrations
          </Link>
          <Typography color="text.primary">All</Typography>
        </Breadcrumbs>

        <Typography variant="h2" gutterBottom component="div">
          Integrations
        </Typography>

        <Box sx={{ my: 3, ml: 1 }}>
          <Typography variant="h4">Add an Integration</Typography>
          <AddIntegrations
            user={user}
            supportedIntegrations={SupportedIntegrations}
          />
        </Box>

        <Divider sx={{ width: '950px' }} />

        <Box sx={{ my: 3, ml: 1 }}>
          <Typography variant="h4">Connected Integrations</Typography>
          <ConnectedIntegrations user={user} />
        </Box>
      </Box>
    </Layout>
  );
};

export default IntegrationsPage;
