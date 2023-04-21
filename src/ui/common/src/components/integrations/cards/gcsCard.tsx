import Box from '@mui/material/Box';
import React from 'react';

import { GCSConfig, Integration } from '../../../utils/integrations';
import { CardTextEntry } from './text';

type Props = {
  integration: Integration;
};

// This should be set to the minimum width required to display the longest category name on the card.
const categoryWidth = '50px';

export const GCSCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as GCSConfig;

  return (
    <Box>
      <CardTextEntry
        category="Bucket: "
        value={config.bucket}
        categoryWidth={categoryWidth}
      />
    </Box>
  );
};
