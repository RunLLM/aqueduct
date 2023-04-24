import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';

import { DatabricksConfig, Integration } from '../../../utils/integrations';

type DatabricksCardProps = {
  integration: Integration;
};

export const DatabricksCard: React.FC<DatabricksCardProps> = ({
  integration,
}) => {
  const config = integration.config as DatabricksConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <Typography variant="body2">
        <strong>Workspace URL: </strong>
        {config.workspace_url}
      </Typography>
      <Typography variant="body2">
        <strong>Access Token: </strong>
        {config.access_token}
      </Typography>
      <Typography variant="body2">
        <strong>Instance Pool ID: </strong>
        {config.instance_pool_id}
      </Typography>
    </Box>
  );
};
