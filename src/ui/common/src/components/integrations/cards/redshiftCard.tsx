import Box from '@mui/material/Box';
import React from 'react';

import { Integration, RedshiftConfig } from '../../../utils/integrations';
import { CardTextEntry } from './text';

type Props = {
  integration: Integration;
};

// This should be set to the minimum width required to display the longest category name on the card.
const categoryWidth = '70px';

export const RedshiftCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as RedshiftConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <CardTextEntry
        category="Host: "
        value={config.host}
        categoryWidth={categoryWidth}
      />

      <CardTextEntry
        category="User: "
        value={config.username}
        categoryWidth={categoryWidth}
      />

      <CardTextEntry
        category="Database: "
        value={config.database}
        categoryWidth={categoryWidth}
      />
    </Box>
  );
};
