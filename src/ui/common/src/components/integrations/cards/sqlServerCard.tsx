import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import { Integration, SqlServerConfig } from '../../../utils/integrations';
import React from 'react';

type Props = {
  integration: Integration;
};

export const SqlServerCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as SqlServerConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <Typography variant="body1">
        <strong>Host: </strong>
        {config.host}
      </Typography>
      <Typography variant="body1">
        <strong>Port: </strong>
        {config.port}
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
