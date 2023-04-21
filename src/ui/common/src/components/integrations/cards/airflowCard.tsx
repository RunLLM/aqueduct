import Box from '@mui/material/Box';
import React from 'react';

import { AirflowConfig, Integration } from '../../../utils/integrations';
import { TruncatedText } from './text';

type Props = {
  integration: Integration;
};

export const AirflowCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as AirflowConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <TruncatedText variant="body2">
        <strong>Host: </strong>
        {config.host}
      </TruncatedText>
      <TruncatedText variant="body2">
        <strong>Username: </strong>
        {config.username}
      </TruncatedText>
      <TruncatedText variant="body2">
        <strong>S3 Credentials Path: </strong>
        {config.s3_credentials_path}
      </TruncatedText>
      <TruncatedText variant="body2">
        <strong>S3 Credentials Profile: </strong>
        {config.s3_credentials_profile}
      </TruncatedText>
    </Box>
  );
};
