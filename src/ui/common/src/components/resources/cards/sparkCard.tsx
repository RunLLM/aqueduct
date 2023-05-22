import React from 'react';

import { Resource, SparkConfig } from '../../../utils/resources';
import { ResourceCardText } from './text';

type SparkCardProps = {
  resource: Resource;
};

export const SparkCard: React.FC<SparkCardProps> = ({ resource }) => {
  const config = resource.config as SparkConfig;
  return (
    <ResourceCardText
      labels={['Livy Server URL']}
      values={[config.livy_server_url]}
    />
  );
};

export default SparkCard;
