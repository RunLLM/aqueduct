import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import { EmailConfig, Integration } from '../../../utils/integrations';

type Props = {
  integration: Integration;
};

export const EmailCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as EmailConfig;
  const targets = JSON.parse(config.targets_serialized) as string[];
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <Typography variant="body2">
        <strong>Sender Address: </strong>
        {config.user} on {config.host}:{config.port}
      </Typography>
      <Typography variant="body2">
        {targets.length > 1 ? (
          <strong>Receiver Addresses: </strong>
        ) : (
          <strong>Receiver Address:</strong>
        )}{' '}
        {targets.join(', ')}
      </Typography>
      {config.enabled === 'true' && (
        <Typography variant="body2">
          <strong>Level: </strong>
          {config.level[0].toUpperCase() + config.level.slice(1)}
        </Typography>
      )}
      {config.enabled !== 'true' && (
        <Typography variant="body2">
          By default, this notification does NOT apply to all workflows.
        </Typography>
      )}
    </Box>
  );
};
