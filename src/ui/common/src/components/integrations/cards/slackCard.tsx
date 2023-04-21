import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';

import { Integration, SlackConfig } from '../../../utils/integrations';
import { CardTextEntry } from './text';

type Props = {
  integration: Integration;
};

// This should be set to the minimum width required to display the longest category name on the card.
const categoryWidth = '60px';

export const SlackCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as SlackConfig;
  const channels = JSON.parse(config.channels_serialized) as string[];
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <CardTextEntry
        category={channels.length > 1 ? 'Channels: ' : 'Channel: '}
        value={channels.join(', ')}
        categoryWidth={categoryWidth}
      />

      {config.enabled === 'true' && (
        <CardTextEntry
          category="Level: "
          value={config.level[0].toUpperCase() + config.level.slice(1)}
          categoryWidth={categoryWidth}
        />
      )}

      {config.enabled !== 'true' && (
        <Typography variant="body2" sx={{ fontWeight: 300, marginTop: 1 }}>
          By default, this notification does NOT apply to all workflows.
        </Typography>
      )}
    </Box>
  );
};
