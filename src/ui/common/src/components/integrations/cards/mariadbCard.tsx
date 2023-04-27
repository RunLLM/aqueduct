import React from 'react';

import { Integration, MariaDbConfig } from '../../../utils/integrations';
import { ResourceCardText } from './text';

type Props = {
  integration: Integration;
};

export const MariaDbCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as MariaDbConfig;
  return (
    <ResourceCardText
      labels={['Host', 'User', 'Database']}
      values={[config.host, config.username, config.database]}
    />
  );
};
