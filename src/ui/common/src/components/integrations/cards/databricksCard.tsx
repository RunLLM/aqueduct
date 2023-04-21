import Box from '@mui/material/Box';
import React from 'react';

import { DatabricksConfig, Integration } from '../../../utils/integrations';
import { TruncatedText } from './text';

type DatabricksCardProps = {
  integration: Integration;
};

export const DatabricksCard: React.FC<DatabricksCardProps> = ({
  integration,
}) => {
  const config = integration.config as DatabricksConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <TruncatedText variant="body2">
        <strong>Workspace URL: </strong>
        {config.workspace_url}
      </TruncatedText>
      <TruncatedText variant="body2">
        <strong>Access Token: </strong>
        {config.access_token}
      </TruncatedText>
    </Box>
  );
};
