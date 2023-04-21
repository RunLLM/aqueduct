import Box from '@mui/material/Box';
import React from 'react';

import { EmailConfig, Integration } from '../../../utils/integrations';
import {TruncatedText} from "./truncatedText";

type Props = {
  integration: Integration;
};

export const EmailCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as EmailConfig;
  const targets = JSON.parse(config.targets_serialized) as string[];
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column'}}>
      <TruncatedText variant="body2">
        <strong>Sender Address: </strong>
        {config.user} on {config.host}:{config.port}
      </TruncatedText>
      <TruncatedText variant="body2">
        {targets.length > 1 ? (
          <strong>Receiver Addresses: </strong>
        ) : (
          <strong>Receiver Address:</strong>
        )}{' '}
        {targets.join(', ')}
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
