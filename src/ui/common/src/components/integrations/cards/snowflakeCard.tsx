import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';

import { Integration, SnowflakeConfig } from '../../../utils/integrations';

type Props = {
  integration: Integration;
};

export const SnowflakeCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as SnowflakeConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <Typography variant="body2">
        <strong>Account Identifier: </strong>
        {config.account_identifier}
      </Typography>
      <Typography variant="body2">
        <strong>Warehouse: </strong>
        {config.warehouse}
      </Typography>
      <Typography variant="body2">
        <strong>User: </strong>
        {config.username}
      </Typography>
      <Typography variant="body2">
        <strong>Database: </strong>
        {config.database}
      </Typography>
      <Typography variant="body2">
        <strong>Schema: </strong>
        {config.schema ? config.schema : 'public'}
      </Typography>
      {config.role && (
        <Typography variant="body2">
          <strong>Role: </strong>
          {config.role}
        </Typography>
      )}
    </Box>
  );
};
