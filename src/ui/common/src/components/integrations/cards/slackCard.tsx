import Box from '@mui/material/Box';
import React from 'react';

import { Integration, SlackConfig } from '../../../utils/integrations';
import { TruncatedText } from './text';

type Props = {
  integration: Integration;
};

export const SlackCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as SlackConfig;
  const channels = JSON.parse(config.channels_serialized) as string[];
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <TruncatedText variant="body2">
        {channels.length > 1 ? (
          <strong>Channel: </strong>
        ) : (
          <strong>Channels:</strong>
        )}{' '}
        {channels.join(', ')}
      </TruncatedText>
      {config.enabled === 'true' && (
        <TruncatedText variant="body2">
          <strong>Level: </strong>
          {config.level[0].toUpperCase() + config.level.slice(1)}
        </TruncatedText>
      )}
      {config.enabled !== 'true' && (
        <TruncatedText variant="body2">
          By default, this notification does NOT apply to all workflows.
        </TruncatedText>
      )}
    </Box>
  );
};
