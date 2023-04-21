import Box from '@mui/material/Box';
import React from 'react';

import { DatabricksConfig, Integration } from '../../../utils/integrations';
import { CardTextEntry } from './text';

type DatabricksCardProps = {
  integration: Integration;
};

// This should be set to the minimum width required to display the longest category name on the card.
const categoryWidth = '80px';

export const DatabricksCard: React.FC<DatabricksCardProps> = ({
  integration,
}) => {
  const config = integration.config as DatabricksConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <CardTextEntry
        category="Workspace: "
        value={config.workspace_url}
        categoryWidth={categoryWidth}
      />
    </Box>
  );
};
