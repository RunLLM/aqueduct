import Box from '@mui/material/Box';
import React from 'react';

import { AWSConfig, Integration } from '../../../utils/integrations';
import { TruncatedText } from './text';

type AWSCardProps = {
  integration: Integration;
};

export const AWSCard: React.FC<AWSCardProps> = ({ integration }) => {
  const config = integration.config as AWSConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      {config.region && (
        <TruncatedText variant="body2">
          <strong>Region: </strong>
          {config.region}
        </TruncatedText>
      )}
      {config.config_file_path && (
        <TruncatedText variant="body2">
          <strong>Credential File Path: </strong>
          {config.config_file_path}
        </TruncatedText>
      )}
      {config.config_file_profile && (
        <TruncatedText variant="body2">
          <strong>Profile: </strong>
          {config.config_file_profile}
        </TruncatedText>
      )}
    </Box>
  );
};

export default AWSCard;
