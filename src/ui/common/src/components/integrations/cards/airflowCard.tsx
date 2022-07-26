import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import { AirflowConfig, Integration } from '../../../utils/integrations';

type Props = {
  integration: Integration;
};

export const AirflowCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as AirflowConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <Typography variant="body1">
        <strong>Host: </strong>
        {config.host}
      </Typography>
      <Typography variant="body1">
        <strong>Username: </strong>
        {config.username}
      </Typography>
      <Typography variant="body1">
        <strong>S3 Credentials Path: </strong>
        {config.s3_credentials_path}
      </Typography>
      <Typography variant="body1">
        <strong>S3 Credentials Profile: </strong>
        {config.s3_credentials_profile}
      </Typography>
    </Box>
  );
};
