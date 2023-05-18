import React from 'react';

import { GCSConfig, Integration } from '../../../utils/resources';
import { ResourceCardText } from './text';

type Props = {
  resource: Integration;
};

export const GCSCard: React.FC<Props> = ({ resource }) => {
  const config = resource.config as GCSConfig;

  return <ResourceCardText labels={['Bucket']} values={[config.bucket]} />;
};
