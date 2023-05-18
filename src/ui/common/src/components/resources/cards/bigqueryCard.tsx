import React from 'react';

import { BigQueryConfig, Integration } from '../../../utils/resources';
import { ResourceCardText } from './text';

type Props = {
  resource: Integration;
};

export const BigQueryCard: React.FC<Props> = ({ resource }) => {
  const config = resource.config as BigQueryConfig;
  return (
    <ResourceCardText labels={['Project ID']} values={[config.project_id]} />
  );
};
