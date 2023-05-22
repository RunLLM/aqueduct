import React from 'react';

import { BigQueryConfig, Resource } from '../../../utils/resources';
import { ResourceCardText } from './text';

type Props = {
  resource: Resource;
};

export const BigQueryCard: React.FC<Props> = ({ resource }) => {
  const config = resource.config as BigQueryConfig;
  return (
    <ResourceCardText labels={['Project ID']} values={[config.project_id]} />
  );
};
