import Box from '@mui/material/Box';
import React from 'react';

import { AWSConfig, Integration } from '../../../utils/integrations';
import { ResourceCardText } from './text';

type AWSCardProps = {
  integration: Integration;
};

export const AWSCard: React.FC<AWSCardProps> = ({ integration }) => {
  const config = integration.config as AWSConfig;

  const labels = [];
  const values = [];

  if (config.region) {
    labels.push('Region');
    values.push(config.region);
  }

  if (config.config_file_path) {
    labels.push('Credential File Path');
    values.push(config.config_file_path);
  }

  if (config.config_file_profile) {
    labels.push('Profile');
    values.push(config.config_file_profile);
  }

  return (
    <Box>
      <ResourceCardText labels={labels} values={values} />
      <Box
        sx={{
          textAlign: 'left',
        }}
      >
        <Typography variant="caption" sx={{ fontWeight: 300 }}>
          Managed by Aqueduct on AWS
        </Typography>
      </Box>
    </Box>
  );
};

export default AWSCard;
