import Box from '@mui/material/Box';
import React from 'react';

import { EmailConfig, Integration } from '../../../utils/integrations';
import { ResourceCardText, TruncatedText } from './text';

type Props = {
  integration: Integration;
};

export const EmailCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as EmailConfig;
  const targets = JSON.parse(config.targets_serialized) as string[];

  const labels = [
    targets.length > 1 ? 'Receiver Addresses' : 'Receiver Address',
  ];
  const values = [targets.join(', ')];

  if (config.enabled === 'true') {
    labels.push('Level');
    values.push(config.level[0].toUpperCase() + config.level.slice(1));
  }

  return (
    <Box>
      <ResourceCardText labels={labels} values={values} />

      {config.enabled !== 'true' && (
        <TruncatedText variant="body2">
          This notification does <strong style={{ fontWeight: 'bold' }}>not</strong> apply to all workflows.
        </TruncatedText>
      )}
    </Box>
  );
};
