import { Checkbox, FormControlLabel } from '@mui/material';
import Box from '@mui/material/Box';
import React, { useState } from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import {
  KubernetesConfig,
  ResourceDialogProps,
} from '../../../utils/resources';
import { ResourceTextInputField } from './ResourceTextInputField';

const Placeholders: KubernetesConfig = {
  kubeconfig_path: '/home/ubuntu/.kube/config',
  cluster_name: 'aqueduct',
  use_same_cluster: 'false',
};

interface KuberentesDialogProps extends ResourceDialogProps<KubernetesConfig> {
  inK8sCluster: boolean;
}

export const KubernetesDialog: React.FC<KuberentesDialogProps> = ({
  resourceToEdit,
  inK8sCluster = false,
}) => {
  const { register, setValue } = useFormContext();
  const editMode = !!resourceToEdit;
  if (resourceToEdit) {
    Object.entries(resourceToEdit).forEach(([k, v]) => {
      register(k, { value: v });
    });
  }

  const initialUseSameCluster = resourceToEdit?.use_same_cluster ?? 'false';
  const [useSameCluster, setUseSameCluster] = useState(initialUseSameCluster);

  return (
    <Box sx={{ mt: 2 }}>
      {inK8sCluster && (
        <FormControlLabel
          label="Use the same Kubernetes cluster that the server is running on."
          control={
            <Checkbox
              checked={useSameCluster === 'true'}
              onChange={(event) => {
                const value = event.target.checked ? 'true' : 'false';
                setUseSameCluster(value);
                setValue('use_same_cluster', value);
              }}
            />
          }
        />
      )}

      <ResourceTextInputField
        name="kubeconfig_path"
        spellCheck={false}
        required={!(useSameCluster === 'true')}
        label="Kubernetes Config Path*"
        description="The path to the kubeconfig file."
        placeholder={Placeholders.kubeconfig_path}
        onChange={(event) => setValue('kubeconfig_path', event.target.value)}
        disabled={useSameCluster === 'true' || editMode}
      />

      <ResourceTextInputField
        name="cluster_name"
        spellCheck={false}
        required={!(useSameCluster === 'true')}
        label="Cluster Name*"
        description="The name of the cluster that will be used."
        placeholder={Placeholders.cluster_name}
        onChange={(event) => setValue('cluster_name', event.target.value)}
        disabled={useSameCluster === 'true' || editMode}
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
