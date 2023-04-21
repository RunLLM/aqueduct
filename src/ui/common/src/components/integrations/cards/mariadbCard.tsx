import Box from '@mui/material/Box';
import React from 'react';

import { Integration, MariaDbConfig } from '../../../utils/integrations';
import { TruncatedText } from './text';

type Props = {
  integration: Integration;
};

export const MariaDbCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as MariaDbConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <TruncatedText variant="body2">
        <strong>Host: </strong>
        {config.host}
      </TruncatedText>
      <TruncatedText variant="body2">
        <strong>Port: </strong>
        {config.port}
      </TruncatedText>
      <TruncatedText variant="body2">
        <strong>User: </strong>
        {config.username}
      </TruncatedText>
      <TruncatedText variant="body2">
        <strong>Database: </strong>
        {config.database}
      </TruncatedText>
    </Box>
  );
};
