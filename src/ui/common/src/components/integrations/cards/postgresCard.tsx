import React from 'react';

import { Integration, PostgresConfig } from '../../../utils/integrations';
import { ResourceCardText } from './text';

type Props = {
  integration: Integration;
};

export const PostgresCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as PostgresConfig;
  return (
    <ResourceCardText
      labels={['Host', 'User', 'Database']}
      values={[config.host, config.username, config.database]}
    />
  );
};
