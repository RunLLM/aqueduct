import React from 'react';

import { Integration, SparkConfig } from '../../../utils/resources';
import { ResourceCardText } from './text';

type SparkCardProps = {
  resource: Integration;
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
