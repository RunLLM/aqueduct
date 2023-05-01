import React from 'react';

import { GCSConfig, Integration } from '../../../utils/integrations';
import { ResourceCardText } from './text';

type Props = {
  integration: Integration;
};

export const GCSCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as GCSConfig;

  return <ResourceCardText labels={['Bucket']} values={[config.bucket]} />;
};
