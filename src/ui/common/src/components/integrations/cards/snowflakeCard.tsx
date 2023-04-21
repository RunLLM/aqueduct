import Box from '@mui/material/Box';
import React from 'react';

import { Integration, SnowflakeConfig } from '../../../utils/integrations';
import { TruncatedText } from './text';

type Props = {
  integration: Integration;
};

export const SnowflakeCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as SnowflakeConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <TruncatedText variant="body2">
        <strong>Account Identifier: </strong>
        {config.account_identifier}
      </TruncatedText>
      <TruncatedText variant="body2">
        <strong>Warehouse: </strong>
        {config.warehouse}
      </TruncatedText>
      <TruncatedText variant="body2">
        <strong>User: </strong>
        {config.username}
      </TruncatedText>
      <TruncatedText variant="body2">
        <strong>Database: </strong>
        {config.database}
      </TruncatedText>
      <TruncatedText variant="body2">
        <strong>Schema: </strong>
        {config.schema ? config.schema : 'public'}
      </TruncatedText>
      {config.role && (
        <TruncatedText variant="body2">
          <strong>Role: </strong>
          {config.role}
        </TruncatedText>
      )}
    </Box>
  );
};
