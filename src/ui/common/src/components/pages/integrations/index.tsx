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
          <ConnectedIntegrations user={user} />
        </Box>
      </Box>
    </Layout>
  );
};

export default IntegrationsPage;
