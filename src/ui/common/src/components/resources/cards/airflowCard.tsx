import React from 'react';

import { AirflowConfig, Resource } from '../../../utils/resources';
import { ResourceCardText } from './text';

type Props = {
  resource: Resource;
};

export const AirflowCard: React.FC<Props> = ({ resource }) => {
  const config = resource.config as AirflowConfig;
  return (
    <ResourceCardText
      labels={['Host', 'Username']}
      values={[config.host, config.username]}
    />
  );
};
