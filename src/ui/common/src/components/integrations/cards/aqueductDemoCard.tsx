import { Alert } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';

import { Integration } from '../../../utils/integrations';

type Props = {
  integration: Integration;
};

export const AqueductDemoCard: React.FC<Props> = ({ integration }) => {
  if (!integration.validated) {
    return (
      <Box sx={{ my: 1 }}>
        <Alert severity="info">
          We are working on spinning up a demo database for you!
        </Alert>
      </Box>
    );
  }
  return null;
};
