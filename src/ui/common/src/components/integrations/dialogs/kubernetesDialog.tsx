import { Checkbox, FormControlLabel } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';
import { useEffect, useState } from 'react';
import { useFormContext } from 'react-hook-form';

import {
  IntegrationDialogProps,
  KubernetesConfig,
} from '../../../utils/integrations';
import { apiAddress } from '../../hooks/useAqueductConsts';
import useUser from '../../hooks/useUser';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: KubernetesConfig = {
  kubeconfig_path: '/home/ubuntu/.kube/config',
  cluster_name: 'aqueduct',
  use_same_cluster: 'false',
};

export const KubernetesDialog: React.FC<IntegrationDialogProps> = ({
  editMode = false,
}) => {
  const { user, loading } = useUser();

  console.log('loading: ', loading);

  const { register, setValue, getValues } = useFormContext();
  const use_same_cluster = getValues('use_same_cluster');

  register('use_same_cluster');

  useEffect(() => {
    setValue('use_same_cluster', 'false');
  }, []);

  const [inK8sCluster, setInK8sCluster] = useState(false);

  // TODO: Move this route over to RTK query
  useEffect(() => {
    const fetchEnvironment = async () => {
      const environmentResponse = await fetch(`${apiAddress}/api/environment`, {
        method: 'GET',
        headers: {
          'api-key': user.apiKey,
        },
      });

      const responseBody = await environmentResponse.json();
      setInK8sCluster(responseBody['inK8sCluster']);
    };

    if (user) {
      fetchEnvironment().catch(console.error);
    }
  }, [user]);

  return (
    <Box sx={{ mt: 2 }}>
      {inK8sCluster && (
        <FormControlLabel
          label="Use the same Kubernetes cluster that the server is running on."
          control={
            <Checkbox
              checked={use_same_cluster === 'true'}
              onChange={(event) =>
                setValue(
                  'use_same_cluster',
                  event.target.checked ? 'true' : 'false'
                )
              }
            />
          }
        />
      )}

      <IntegrationTextInputField
        name="kubeconfig_path"
        spellCheck={false}
        required={!(use_same_cluster === 'true')}
        label="Kubernetes Config Path*"
        description="The path to the kubeconfig file."
        placeholder={Placeholders.kubeconfig_path}
        onChange={(event) => setValue('kubeconfig_path', event.target.value)}
        disabled={use_same_cluster === 'true'}
      />

      <IntegrationTextInputField
        name="cluster_name"
        spellCheck={false}
        required={!(use_same_cluster === 'true')}
        label="Cluster Name*"
        description="The name of the cluster that will be used."
        placeholder={Placeholders.cluster_name}
        onChange={(event) => setValue('cluster_name', event.target.value)}
        disabled={use_same_cluster === 'true'}
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
