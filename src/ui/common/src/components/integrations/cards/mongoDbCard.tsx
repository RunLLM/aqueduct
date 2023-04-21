import Box from '@mui/material/Box';
import React from 'react';

import { Integration, MongoDBConfig } from '../../../utils/integrations';
import { CardTextEntry } from './text';

type Props = {
  integration: Integration;
};

// This should be set to the minimum width required to display the longest category name on the card.
const categoryWidth = '70px';

export const MongoDBCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as MongoDBConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <CardTextEntry
        category="Database: "
        value={config.database}
        categoryWidth={categoryWidth}
      />
    </Box>
  );
};
