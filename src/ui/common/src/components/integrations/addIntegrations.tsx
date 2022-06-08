import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import React, { useState } from 'react';

import UserProfile from '../../utils/auth';
import { Service, ServiceInfoMap } from '../../utils/integrations';
import { IntegrationDialog } from './dialogs/dialog';

type Props = {
  user: UserProfile;
  supportedIntegrations: ServiceInfoMap;
};

const AddIntegrations: React.FC<Props> = ({
  user,
  supportedIntegrations,
}) => {
  return (
    <Box sx={{ maxWidth: '950px' }}>
      <Grid container spacing={2} sx={{ my: '20px', width: '100%' }}>
        {Object.entries(supportedIntegrations)
          .filter(([svc]) => svc !== 'Aqueduct Demo')
          .map(([svc, integration]) => {
            const service = svc as Service;
            const [showDialog, setShowDialog] = useState(false);

            const iconWrapper = (
              <Grid item>
                <Box
                  onClick={() => setShowDialog(integration.activated)}
                  sx={{
                    px: 2,
                    py: 2,
                    mx: 2,
                    borderRadius: 2,
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
                    height="85px"
                    style={{ opacity: integration.activated ? 1.0 : 0.3 }}
                  />
                </Box>
              </Grid>
            );

            return (
              <Box key={service}>
                {iconWrapper}
                {showDialog && (
                  <IntegrationDialog
                    user={user}
                    service={service}
                    onCloseDialog={() => setShowDialog(false)}
                  />
                )}
              </Box>
            );
          })}
      </Grid>
    </Box>
  );
};

export default AddIntegrations;
