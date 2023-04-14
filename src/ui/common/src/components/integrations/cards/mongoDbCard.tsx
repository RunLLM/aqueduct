import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';

import { Integration, MongoDBConfig } from '../../../utils/integrations';

type Props = {
  integration: Integration;
};

export const MongoDBCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as MongoDBConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <Typography variant="body2">
        <strong>URI: </strong>
        ********
      </Typography>
      <Typography variant="body2">
        <strong>Database: </strong>
        {config.database}
      </Typography>
    </Box>
  );
};
