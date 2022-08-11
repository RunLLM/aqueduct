import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import { KubernetesConfig, Integration } from '../../../utils/integrations';

type Props = {
  integration: Integration;
};

export const KubernetesCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as KubernetesConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <Typography variant="body1">
        <strong>Kubernetes Config Path: </strong>
        {config.kube_config_path}
      </Typography>
    </Box>
  );
};
