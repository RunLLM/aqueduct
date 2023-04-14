import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import { GCSConfig, Integration } from '../../../utils/integrations';

type Props = {
  integration: Integration;
};

export const GCSCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as GCSConfig;

  return (
    <Box>
      <Typography variant="body2">
        <strong>Bucket: </strong>
        {config.bucket}
      </Typography>
    </Box>
  );
};
