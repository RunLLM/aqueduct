import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import Divider from '@mui/material/Divider';
import Head from 'next/head';
import React from 'react';

import UserProfile from '../../../utils/auth';
import { SupportedIntegrations } from '../../../utils/integrations';
import { AddIntegrations } from '../../integrations/addIntegrations';
import { ConnectedIntegrations } from '../../integrations/connectedIntegrations';
import DefaultLayout from '../../layouts/default';

type Props = {
  user: UserProfile;
};

const IntegrationsPage: React.FC<Props> = ({ user }) => {
  return (
    <DefaultLayout user={user}>
      <Head>
        <title> Integrations | Aqueduct </title>
      </Head>

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
    </DefaultLayout>
  );
};

export default IntegrationsPage;
