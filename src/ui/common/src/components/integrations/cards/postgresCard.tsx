import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import { Integration, PostgresConfig } from '../../../utils/integrations';

type Props = {
  integration: Integration;
};

export const PostgresCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as PostgresConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <Typography variant="body1">
        <strong>Host: </strong>
        {config.host}
      </Typography>
      <Typography variant="body1">
        <strong>User: </strong>
        {config.username}
      </Typography>
      <Typography variant="body1">
        <strong>Database: </strong>
        {config.database}
      </Typography>
    </Box>
  );
};
