import React from 'react';

import { DatabricksConfig, Integration } from '../../../utils/integrations';
import { ResourceCardText } from './text';

type DatabricksCardProps = {
  integration: Integration;
};

export const DatabricksCard: React.FC<DatabricksCardProps> = ({
  integration,
}) => {
  const config = integration.config as DatabricksConfig;
  return (
    <ResourceCardText labels={['Workspace']} values={[config.workspace_url]} />
  );
};
