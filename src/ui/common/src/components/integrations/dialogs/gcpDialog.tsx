import React from 'react';
import * as Yup from 'yup';

import { IntegrationDialogProps } from '../../../utils/integrations';
import { IntegrationTextInputField } from './IntegrationTextInputField';

// Placeholder component for the GCP dialog.
export const GCPDialog: React.FC<IntegrationDialogProps> = ({
  editMode = false,
}) => {
  return (
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
  );
};

export function getGCPValidationSchema() {
  return Yup.object().shape({
    cluster_name: Yup.string().required('Please enter a cluster name'),
  });
}

export default GCPDialog;
