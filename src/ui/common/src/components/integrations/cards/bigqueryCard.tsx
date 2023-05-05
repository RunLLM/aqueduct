import React from 'react';

import { BigQueryConfig, Integration } from '../../../utils/integrations';
import { ResourceCardText } from './text';

type Props = {
  integration: Integration;
};

export const BigQueryCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as BigQueryConfig;
  return (
    <ResourceCardText labels={['Project ID']} values={[config.project_id]} />
  );
};
