import React from 'react';

import { Integration, KubernetesConfig } from '../../../utils/resources';
import { ResourceCardText } from './text';

type Props = {
  resource: Integration;
};

export const KubernetesCard: React.FC<Props> = ({ resource }) => {
  const config = resource.config as KubernetesConfig;
  return (
    <ResourceCardText
      labels={['Kube Config', 'Cluster Name']}
      values={[config.kubeconfig_path, config.cluster_name]}
    />
  );
};
