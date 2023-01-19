import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import { EmailConfig, Integration } from '../../../utils/integrations';

type Props = {
  integration: Integration;
};

export const EmailCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as EmailConfig;
  const target = (JSON.parse(config.targets_serialized) as string[])[0];
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <Typography variant="body2">
        <strong>Host: </strong>
        {config.host}
      </Typography>
      <Typography variant="body2">
        <strong>Port: </strong>
        {config.port}
      </Typography>
      <Typography variant="body2">
        <strong>Sender Address: </strong>
        {config.user}
      </Typography>
      <Typography variant="body2" color={!!target ? 'black' : 'gray700'}>
        <strong>Receiver Address: </strong>
        {target ?? 'Not specified'}
      </Typography>
      <Typography variant="body2" color={!!target ? 'black' : 'gray700'}>
        <strong>Level: </strong>
        {config.level}
      </Typography>
    </Box>
  );
};
