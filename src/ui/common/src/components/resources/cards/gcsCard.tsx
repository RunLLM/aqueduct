import React from 'react';

import { GCSConfig, Resource } from '../../../utils/resources';
import { ResourceCardText } from './text';

type Props = {
  resource: Resource;
};

export const GCSCard: React.FC<Props> = ({ resource }) => {
  const config = resource.config as GCSConfig;

  return <ResourceCardText labels={['Bucket']} values={[config.bucket]} />;
};
