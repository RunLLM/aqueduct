import Box from '@mui/material/Box';
import React from 'react';

import { EmailConfig, Integration } from '../../../utils/integrations';
import { CardTextEntry, TruncatedText } from './text';

type Props = {
  integration: Integration;
};

// This should be set to the minimum width required to display the longest category name on the card.
const categoryWidth = '120px';

export const EmailCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as EmailConfig;
  const targets = JSON.parse(config.targets_serialized) as string[];
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <CardTextEntry
        category={
          targets.length > 1 ? 'Receiver Addresses: ' : 'Receiver Address: '
        }
        value={targets.join(', ')}
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
        <TruncatedText variant="body2">
          By default, this notification does NOT apply to all workflows.
        </TruncatedText>
      )}
    </Box>
  );
};
