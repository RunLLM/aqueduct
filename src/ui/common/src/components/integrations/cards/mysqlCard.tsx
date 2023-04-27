import React from 'react';

import { Integration, MySqlConfig } from '../../../utils/integrations';
import { ResourceCardText } from './text';

type Props = {
  integration: Integration;
};

export const MySqlCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as MySqlConfig;
  return (
    <ResourceCardText
      labels={['Host', 'User', 'Database']}
      values={[config.host, config.username, config.database]}
    />
  );
};
