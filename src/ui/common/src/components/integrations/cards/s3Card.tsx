import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import { Integration } from '../../../utils/integrations';
import { S3Config } from '../../../utils/workflows';
import StorageConfigurationDisplay from '../StorageConfiguration';

type Props = {
  integration: Integration;
};

export const S3Card: React.FC<Props> = ({ integration }) => {
  const config = integration.config as S3Config;

  return (
    <Box>
      <Typography variant="body2">
        <strong>Bucket: </strong>
        {config.bucket}
      </Typography>
      <Typography variant="body2">
        <strong>Region: </strong>
        {config.region}
      </Typography>
      <StorageConfigurationDisplay integrationName="s3" />
    </Box>
  );
};
