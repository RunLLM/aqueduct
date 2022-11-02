import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';

import { Integration, MongoDbConfig } from '../../../utils/integrations';

type Props = {
  integration: Integration;
};

export const MongoDbCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as MongoDbConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <Typography variant="body1">
        <strong>Uri: </strong>
        ********
      </Typography>
      <Typography variant="body1">
        <strong>Database: </strong>
        {config.database}
      </Typography>
    </Box>
  );
};
