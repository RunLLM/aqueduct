import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import { Integration } from '../../../utils/integrations';
import { S3Config } from '../../../utils/workflows';

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
      {config.root_dir?.length > 0 && (
        <Typography variant="body2">
          <strong>Root Directory: </strong>
          {config.root_dir}
        </Typography>
      )}
      <Typography variant="body2">
        <strong>Region: </strong>
        {config.region}
      </Typography>
    </Box>
  );
};
