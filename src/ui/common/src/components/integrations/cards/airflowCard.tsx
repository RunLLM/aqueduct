import Box from '@mui/material/Box';
import React from 'react';

import { AirflowConfig, Integration } from '../../../utils/integrations';
import { CardTextEntry } from './text';

type Props = {
  integration: Integration;
};

// This should be set to the minimum width required to display the longest category name on the card.
const categoryWidth = '75px';

export const AirflowCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as AirflowConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <CardTextEntry
        category="Host: "
        value={config.host}
        categoryWidth={categoryWidth}
      />

      <CardTextEntry
        category="Username: "
        value={config.username}
        categoryWidth={categoryWidth}
      />
    </Box>
  );
};
