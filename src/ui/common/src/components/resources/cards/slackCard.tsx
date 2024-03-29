import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';

import { Resource, SlackConfig } from '../../../utils/resources';
import { ResourceCardText } from './text';

type Props = {
  resource: Resource;
};

export const SlackCard: React.FC<Props> = ({ resource }) => {
  const config = resource.config as SlackConfig;
  const channels = JSON.parse(config.channels_serialized) as string[];

  const labels = [channels.length > 1 ? 'Channels' : 'Channel'];
  const values = [channels.join(', ')];

  if (config.enabled === 'true') {
    labels.push('Level');
    values.push(config.level[0].toUpperCase() + config.level.slice(1));
  }

  return (
    <Box>
      <ResourceCardText labels={labels} values={values} />

      {config.enabled !== 'true' && (
        <Typography variant="body2" sx={{ fontWeight: 300, marginTop: 1 }}>
          This notification does{' '}
          <strong style={{ fontWeight: 'bold' }}>not</strong> apply to all
          workflows.
        </Typography>
      )}
    </Box>
  );
};
