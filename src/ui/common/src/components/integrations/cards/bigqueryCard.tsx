import Box from '@mui/material/Box';
import React from 'react';

import { BigQueryConfig, Integration } from '../../../utils/integrations';
import { TruncatedText } from './text';

type Props = {
  integration: Integration;
};

export const BigQueryCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as BigQueryConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <TruncatedText variant="body2">
        <strong>Project ID: </strong>
        {config.project_id}
      </TruncatedText>
    </Box>
  );
};
