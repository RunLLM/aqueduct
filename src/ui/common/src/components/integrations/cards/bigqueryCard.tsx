import Box from '@mui/material/Box';
import React from 'react';

import { BigQueryConfig, Integration } from '../../../utils/integrations';
import { CardTextEntry } from './text';

type Props = {
  integration: Integration;
};

// This should be set to the minimum width required to display the longest category name on the card.
const categoryWidth = '70px';

export const BigQueryCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as BigQueryConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <CardTextEntry
        category="Project ID: "
        value={config.project_id}
        categoryWidth={categoryWidth}
      />
    </Box>
  );
};
