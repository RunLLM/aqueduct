import React from 'react';

import { Integration, SnowflakeConfig } from '../../../utils/integrations';
import { ResourceCardText } from './text';

type Props = {
  integration: Integration;
};

export const SnowflakeCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as SnowflakeConfig;
  return (
    <ResourceCardText
      labels={['Account ID', 'Database', 'User']}
      values={[config.account_identifier, config.database, config.username]}
    />
  );
};
