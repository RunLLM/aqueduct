import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';

import { AWSConfig, Integration } from '../../../utils/integrations';

type AWSCardProps = {
  integration: Integration;
};

export const AWSCard: React.FC<AWSCardProps> = ({ integration }) => {
  const config = integration.config as AWSConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <Typography variant="body2">
        <strong>TBD: </strong>
        {config.region}
      </Typography>
    </Box>
  );
};

export default AWSCard;
