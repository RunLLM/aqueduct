import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';

import { BigQueryConfig, Integration } from '../../../utils/integrations';

type Props = {
  integration: Integration;
};

export const BigQueryCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as BigQueryConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <Typography variant="body2">
        <strong>Project ID: </strong>
        {config.project_id}
      </Typography>
    </Box>
  );
};
