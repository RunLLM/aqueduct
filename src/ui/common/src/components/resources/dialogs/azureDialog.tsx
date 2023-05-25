import React from 'react';
import * as Yup from 'yup';

import { ResourceDialogProps } from '../../../utils/resources';
import { ResourceTextInputField } from './ResourceTextInputField';

// Placeholder component for the Azure dialog.
export const AzureDialog: React.FC<ResourceDialogProps> = ({
  editMode = false,
}) => {
  return (
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
  );
};

export function getAzureValidationSchema() {
  return Yup.object().shape({
    cluster_name: Yup.string().required('Please enter a cluster name'),
  });
}

export default AzureDialog;
