import React from 'react';

import { Integration, RedshiftConfig } from '../../../utils/integrations';
import { ResourceCardText } from './text';

type Props = {
  integration: Integration;
};

export const RedshiftCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as RedshiftConfig;
  return (
    <ResourceCardText
      labels={['Host', 'User', 'Database']}
      values={[config.host, config.username, config.database]}
    />
  );
};
