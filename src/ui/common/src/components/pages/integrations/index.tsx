import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import Divider from '@mui/material/Divider';
import React, { useEffect } from 'react';

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

const IntegrationsPage: React.FC<Props> = ({ user, Layout = DefaultLayout }) => {
  useEffect(() => {
    document.title = 'Integrations | Aqueduct';
  }, []);

  return (
    <Layout user={user}>
      <Box>
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
