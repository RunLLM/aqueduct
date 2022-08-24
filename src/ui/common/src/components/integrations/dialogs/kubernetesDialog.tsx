import Box from '@mui/material/Box';
import React from 'react';

import { KubernetesConfig } from '../../../utils/integrations';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: KubernetesConfig = {
  kubeconfig_path: 'home/ubuntu/.kube/config',
  cluster_name: 'aqueduct',
};

type Props = {
  onUpdateField: (field: keyof KubernetesConfig, value: string) => void;
  value?: KubernetesConfig;
};

export const KubernetesDialog: React.FC<Props> = ({ onUpdateField, value }) => {
  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Kubernetes Config Path*"
        description="The path to the kubeconfig file."
        placeholder={Placeholders.kubeconfig_path}
        onChange={(event) =>
          onUpdateField('kubeconfig_path', event.target.value)
        }
        value={value?.kubeconfig_path ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Cluster Name*"
        description="The name of the cluster that will be used."
        placeholder={Placeholders.cluster_name}
        onChange={(event) => onUpdateField('cluster_name', event.target.value)}
        value={value?.cluster_name ?? null}
      />
    </Box>
  );
};
