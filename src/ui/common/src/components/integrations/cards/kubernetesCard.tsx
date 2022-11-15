import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import { Integration, KubernetesConfig } from '../../../utils/integrations';

type Props = {
  integration: Integration;
};

export const KubernetesCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as KubernetesConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <Typography variant="body2">
        <strong>Kubernetes Config Path: </strong>
        {config.kubeconfig_path}
      </Typography>
      <Typography variant="body2">
        <strong>Cluster Name: </strong>
        {config.cluster_name}
      </Typography>
    </Box>
  );
};
