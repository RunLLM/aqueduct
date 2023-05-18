import React from 'react';

import { Integration, MongoDBConfig } from '../../../utils/resources';
import { ResourceCardText } from './text';

type Props = {
  resource: Integration;
};

export const MongoDBCard: React.FC<Props> = ({ resource }) => {
  const config = resource.config as MongoDBConfig;
  return <ResourceCardText labels={['Database']} values={[config.database]} />;
};
