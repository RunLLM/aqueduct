import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import React from 'react';
import { useState } from 'react';

import { KubernetesConfig } from '../../../utils/integrations';
import { KubernetesDialog } from './kubernetesDialog';

type Props = {
  onUpdateField: (field: keyof KubernetesConfig, value: string) => void;
  value?: KubernetesConfig;
  apiKey: string;
};

export const KubernetesSelectorDialog: React.FC<Props> = ({ onUpdateField, value, apiKey }) => {
  const [showKubernetesDialog, setShowKubernetesDialog] = useState(false);
  const [showOndemandDialog, setShowOndemandDialog] = useState(false);

  const handleOption1Click = () => {
    setShowKubernetesDialog(true);
  };

  const handleOption2Click = () => {
    setShowOndemandDialog(true);
  };

  return (
    <Box>
      {!showKubernetesDialog && !showOndemandDialog ? (
        <>
          <Button variant="contained" color="primary" onClick={handleOption1Click}>
            I have an existing Kubernetes cluster
          </Button>
          <Button variant="contained" color="primary" onClick={handleOption2Click}>
            Create an on-demand Kubernetes integration
          </Button>
        </>
      ) : null}
      {showKubernetesDialog && (
        <KubernetesDialog onUpdateField={onUpdateField} value={value} apiKey={apiKey} />
      )}
      {/* {showOndemandDialog && (
        <OndemandDialog onUpdateField={onUpdateField} value={value} apiKey={apiKey} onCancel={handleCancel} />
      )} */}
    </Box>
  );
};
