import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';

import { AWSConfig, Integration } from '../../../utils/integrations';

type AWSCardProps = {
  integration: Integration;
};

export const AWSCard: React.FC<AWSCardProps> = ({ integration }) => {
  const config = integration.config as AWSConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      {config.region && (
        <Typography variant="body2">
          <strong>Region: </strong>
          {config.region}
        </Typography>
      )}
      {config.config_file_path && (
        <Typography variant="body2">
          <strong>Credential File Path: </strong>
          {config.config_file_path}
        </Typography>
      )}
      {config.config_file_profile && (
        <Typography variant="body2">
          <strong>Profile: </strong>
          {config.config_file_profile}
        </Typography>
      )}
    </Box>
  );
};

export default AWSCard;
