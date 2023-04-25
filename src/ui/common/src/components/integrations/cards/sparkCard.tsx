import React from 'react';

import { Integration, SparkConfig } from '../../../utils/integrations';
import { ResourceCardText } from './text';

type SparkCardProps = {
  integration: Integration;
};

export const SparkCard: React.FC<SparkCardProps> = ({ integration }) => {
  const config = integration.config as SparkConfig;
  return (
    <ResourceCardText
      labels={['Livy Server URL']}
      values={[config.livy_server_url]}
    />
  );
};

export default SparkCard;
