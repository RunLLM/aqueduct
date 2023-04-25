import React from 'react';

import { Integration, KubernetesConfig } from '../../../utils/integrations';
import { ResourceCardText } from './text';

type Props = {
  integration: Integration;
};

export const KubernetesCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as KubernetesConfig;
  return (
    <ResourceCardText
      labels={['Kube Config', 'Cluster Name']}
      values={[config.kubeconfig_path, config.cluster_name]}
    />
  );
};
