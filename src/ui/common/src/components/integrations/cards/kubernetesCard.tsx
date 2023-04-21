import Box from '@mui/material/Box';
import React from 'react';
import { TruncatedText } from './truncatedText';
import { Integration, KubernetesConfig } from '../../../utils/integrations';

type Props = {
  integration: Integration;
};

export const KubernetesCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as KubernetesConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <TruncatedText variant="body2">
        <strong>Kubernetes Config Path: </strong>
        {config.kubeconfig_path}
      </TruncatedText>
      <TruncatedText variant="body2">
        <strong>Cluster Name: </strong>
        {config.cluster_name}
      </TruncatedText>
    </Box>
  );
};
