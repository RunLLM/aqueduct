import { Checkbox, FormControlLabel } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';
import { useEffect, useState } from 'react';

import { KubernetesConfig } from '../../../utils/integrations';
import { apiAddress } from '../../hooks/useAqueductConsts';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: KubernetesConfig = {
  kubeconfig_path: '/home/ubuntu/.kube/config',
  cluster_name: 'aqueduct',
  use_same_cluster: 'false',
};

type Props = {
  onUpdateField: (field: keyof KubernetesConfig, value: string) => void;
  value?: KubernetesConfig;
  apiKey: string;
};

export const KubernetesDialog: React.FC<Props> = ({
  onUpdateField,
  value,
  apiKey,
}) => {
  const [inK8sCluster, setInK8sCluster] = useState(false);
  useEffect(() => {
    if (!value?.use_same_cluster) {
      onUpdateField('use_same_cluster', 'false');
    }
  }, [apiKey, onUpdateField, value?.use_same_cluster]);

  useEffect(() => {
    const fetchEnvironment = async () => {
      const environmentResponse = await fetch(`${apiAddress}/api/environment`, {
        method: 'GET',
        headers: {
          'api-key': apiKey,
        },
      });

      const responseBody = await environmentResponse.json();
      setInK8sCluster(responseBody['inK8sCluster']);
    };

    fetchEnvironment().catch(console.error);
  }, [apiKey]);

  console.log('rendering kubernetes dialog ...');

  return (
    <Box sx={{ mt: 2 }}>
      {inK8sCluster && (
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
      )}

      <IntegrationTextInputField
        spellCheck={false}
        required={!(value?.use_same_cluster === 'true')}
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
        required={!(value?.use_same_cluster === 'true')}
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

export function isK8sConfigComplete(config: KubernetesConfig): boolean {
  if (config.use_same_cluster !== 'true') {
    return !!config.kubeconfig_path && !!config.cluster_name;
  }

  // If the user configures to run compute from within the same k8s cluster, we don't need parameters above.
  return true;
}
