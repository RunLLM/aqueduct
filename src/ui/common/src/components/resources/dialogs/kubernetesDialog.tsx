import { Checkbox, FormControlLabel } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';
import { useEffect } from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import {
  ResourceDialogProps,
  KubernetesConfig,
} from '../../../utils/integrations';
import { ResourceTextInputField } from './ResourceTextInputField';

const Placeholders: KubernetesConfig = {
  kubeconfig_path: '/home/ubuntu/.kube/config',
  cluster_name: 'aqueduct',
  use_same_cluster: 'false',
};

interface KuberentesDialogProps extends ResourceDialogProps {
  inK8sCluster: boolean;
}

export const KubernetesDialog: React.FC<ResourceDialogProps> = ({
  editMode = false,
  user,
  inK8sCluster = false,
}) => {
  const { register, setValue, getValues } = useFormContext();
  const use_same_cluster = getValues('use_same_cluster');

  useEffect(() => {
    if (!use_same_cluster) {
      setValue('use_same_cluster', 'false');
    }
  }, []);

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

      <ResourceTextInputField
        name="kubeconfig_path"
        spellCheck={false}
        required={!(use_same_cluster === 'true')}
        label="Kubernetes Config Path*"
        description="The path to the kubeconfig file."
        placeholder={Placeholders.kubeconfig_path}
        onChange={(event) => setValue('kubeconfig_path', event.target.value)}
        disabled={use_same_cluster === 'true'}
      />

      <ResourceTextInputField
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

export function getKubernetesValidationSchema() {
  return Yup.object().shape({
    name: Yup.string().required('Please enter a name.'),
    use_same_cluster: Yup.string().transform((value) => {
      if (value === 'true') {
        return 'true';
      }

      return 'false';
    }),
    kubeconfig_path: Yup.string().when('use_same_cluster', {
      is: 'false',
      then: Yup.string().required('Please enter a kubeconfig path'),
    }),
    cluster_name: Yup.string().when('use_same_cluster', {
      is: 'false',
      then: Yup.string().required('Please enter a cluster name'),
    }),
  });
}
