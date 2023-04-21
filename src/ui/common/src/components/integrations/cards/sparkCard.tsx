import Box from '@mui/material/Box';
import React from 'react';

import { Integration, SparkConfig } from '../../../utils/integrations';
import { CardTextEntry } from './text';

type SparkCardProps = {
  integration: Integration;
};

// This should be set to the minimum width required to display the longest category name on the card.
const categoryWidth = '110px';

export const SparkCard: React.FC<SparkCardProps> = ({ integration }) => {
  const config = integration.config as SparkConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <CardTextEntry
        category="Livy Server URL: "
        value={config.livy_server_url}
        categoryWidth={categoryWidth}
      />
    </Box>
  );
};

export default SparkCard;
