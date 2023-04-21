import Box from '@mui/material/Box';
import React from 'react';

import { Integration, KubernetesConfig } from '../../../utils/integrations';
import { CardTextEntry } from './text';

type Props = {
  integration: Integration;
};

// This should be set to the minimum width required to display the longest category name on the card.
const categoryWidth = '100px';

export const KubernetesCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as KubernetesConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <CardTextEntry
        category="Kube Config: "
        value={config.kubeconfig_path}
        categoryWidth={categoryWidth}
      />

      <CardTextEntry
        category="Cluster Name: "
        value={config.cluster_name}
        categoryWidth={categoryWidth}
      />
    </Box>
  );
};
