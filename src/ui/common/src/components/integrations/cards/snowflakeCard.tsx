import Box from '@mui/material/Box';
import React from 'react';

import { Integration, SnowflakeConfig } from '../../../utils/integrations';
import { CardTextEntry } from './text';

type Props = {
  integration: Integration;
};

// This should be set to the minimum width required to display the longest category name on the card.
const categoryWidth = '80px';

export const SnowflakeCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as SnowflakeConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <CardTextEntry
        category="Account ID: "
        value={config.account_identifier}
        categoryWidth={categoryWidth}
      />

      <CardTextEntry
        category="Database: "
        value={config.database}
        categoryWidth={categoryWidth}
      />

      <CardTextEntry
        category="User: "
        value={config.username}
        categoryWidth={categoryWidth}
      />
    </Box>
  );
};
