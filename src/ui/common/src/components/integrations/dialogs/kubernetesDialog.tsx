import Box from '@mui/material/Box';
import React from 'react';
import { Checkbox, FormControlLabel } from '@mui/material';

import { KubernetesConfig } from '../../../utils/integrations';
import { IntegrationTextInputField } from './IntegrationTextInputField';
import { useEffect } from 'react';

const Placeholders: KubernetesConfig = {
  kubeconfig_path: 'home/ubuntu/.kube/config',
  cluster_name: 'aqueduct',
  use_same_cluster: 'false',
};

type Props = {
  onUpdateField: (field: keyof KubernetesConfig, value: string) => void;
  value?: KubernetesConfig;
};

export const KubernetesDialog: React.FC<Props> = ({ onUpdateField, value }) => {
  useEffect(() => {
    if (!value?.use_same_cluster) {
      onUpdateField('use_same_cluster', 'false');
    }
  }, []);

  return (
    <Box sx={{ mt: 2 }}>

      <FormControlLabel
        label="Use the same Kubernetes cluster that the server is running on."
        control={
          <Checkbox
            checked={value?.use_same_cluster === 'true'}
            onChange={(event) =>
              onUpdateField(
                'use_same_cluster',
                event.target.checked ? 'true' : 'false'
              )
            }
          />
        }
      />
      

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
        disabled={value?.use_same_cluster === 'true'}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Cluster Name*"
        description="The name of the cluster that will be used."
        placeholder={Placeholders.cluster_name}
        onChange={(event) => onUpdateField('cluster_name', event.target.value)}
        value={value?.cluster_name ?? null}
        disabled={value?.use_same_cluster === 'true'}
      />
    </Box>
  );
};
