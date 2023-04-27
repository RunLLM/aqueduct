import React from 'react';

import { Integration, MongoDBConfig } from '../../../utils/integrations';
import { ResourceCardText } from './text';

type Props = {
  integration: Integration;
};

export const MongoDBCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as MongoDBConfig;
  return <ResourceCardText labels={['Database']} values={[config.database]} />;
};
