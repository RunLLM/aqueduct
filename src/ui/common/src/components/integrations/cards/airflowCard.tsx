import React from 'react';

import { AirflowConfig, Integration } from '../../../utils/integrations';
import { ResourceCardText } from './text';

type Props = {
  integration: Integration;
};

export const AirflowCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as AirflowConfig;
  return (
    <ResourceCardText
      labels={['Host', 'Username']}
      values={[config.host, config.username]}
    />
  );
};
