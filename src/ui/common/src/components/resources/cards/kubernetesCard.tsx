import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';

import { KubernetesConfig, Resource } from '../../../utils/resources';
import { ResourceCardText } from './text';

type Props = {
  resource: Resource;
};

export const KubernetesCard: React.FC<Props> = ({ resource }) => {
  const config = resource.config as KubernetesConfig;
  return (
    <Box>
      <ResourceCardText
        labels={['Kube Config', 'Cluster Name']}
        values={[config.kubeconfig_path, config.cluster_name]}
      />
      {config.cloud_provider === 'GCP' && (
        <Box
          sx={{
            textAlign: 'left',
          }}
        >
          <Typography variant="caption" sx={{ fontWeight: 300 }}>
            Managed by Aqueduct on GCP
          </Typography>
        </Box>
      )}
    </Box>
  );
};
