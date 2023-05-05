import Box from '@mui/material/Box';
import React from 'react';

import { EmailConfig, Integration } from '../../../utils/integrations';
import { ResourceCardText, TruncatedText } from './text';

type Props = {
  integration: Integration;
  detailedView: boolean;
};

export const EmailCard: React.FC<Props> = ({ integration, detailedView }) => {
  const config = integration.config as EmailConfig;
  const targets = JSON.parse(config.targets_serialized) as string[];

  let labels = [targets.length > 1 ? 'Receiver Addresses' : 'Receiver Address'];
  let values = [targets.join(', ')];

  if (config.enabled === 'true') {
    labels.push('Level');
    values.push(config.level[0].toUpperCase() + config.level.slice(1));
  }

  if (detailedView) {
    labels = labels.concat(labels, ['Host', 'Port', 'User']);
    values = values.concat(values, [config.host, config.port, config.user]);
  }

  return (
    <Box>
      <ResourceCardText labels={labels} values={values} />

      {config.enabled !== 'true' && (
        <TruncatedText variant="body2" sx={{ fontWeight: 300, marginTop: 1 }}>
          This notification does{' '}
          <strong style={{ fontWeight: 'bold' }}>not</strong> apply to all
          workflows.
        </TruncatedText>
      )}
    </Box>
  );
};
