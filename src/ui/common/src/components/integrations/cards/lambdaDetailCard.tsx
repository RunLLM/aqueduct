import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import { Integration, LambdaConfig } from '../../../utils/integrations';

type Props = {
  integration: Integration;
};

export const LambdaDetailCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as LambdaConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <Typography variant="body1">
        <strong>Lambda Role ARN: </strong>
      </Typography>
      <Typography variant="body1">
        {config.role_arn}
      </Typography>
    </Box>
  );
};
