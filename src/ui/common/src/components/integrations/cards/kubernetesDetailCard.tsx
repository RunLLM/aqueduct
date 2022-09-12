import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import { Integration, KubernetesConfig } from '../../../utils/integrations';

type Props = {
  integration: Integration;
};

export const KubernetesDetailCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as KubernetesConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <Typography variant="body1">
        <strong>Kubernetes Config Path: </strong>
      </Typography>
      <Typography variant="body1">
        {config.kubeconfig_path}
      </Typography>
      <Typography variant="body1">
        <strong>Cluster Name: </strong>
      </Typography>
      <Typography variant="body1">
        {config.cluster_name}
      </Typography>
    </Box>
  );
};
