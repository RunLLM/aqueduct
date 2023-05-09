import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle/DialogTitle';
import Typography from '@mui/material/Typography';
import React, { useState } from 'react';
import { useFormContext } from 'react-hook-form';
import { useDispatch } from 'react-redux';
import * as Yup from 'yup';

import { handleConnectToNewIntegration } from '../../../reducers/integration';
import { AppDispatch } from '../../../stores/store';
import {
  IntegrationDialogProps,
  SupportedIntegrations,
} from '../../../utils/integrations';
import useUser from '../../hooks/useUser';
import IntegrationLogo from '../logo';
import { AWSDialog } from './awsDialog';
import { DialogActionButtons, DialogHeader } from './dialog';
import { KubernetesDialog } from './kubernetesDialog';

export const OnDemandKubernetesDialog: React.FC<IntegrationDialogProps> = ({
  editMode = false,
  disabled,
  loading,
  onCloseDialog,
}) => {
  const [currentStep, setCurrentStep] = useState('INITIAL');

  // INITIAL step is when user is choosing to connect to their own or aqueduct cluster.
  // REGULAR_K8S step is when user is connecting to their own cluster.
  // ONDEMAND_K8S step is when user is connecting to aqueduct cluster.
  // ONDEMAND_K8S_AWS step is when user is connecting to aqueduct cluster on AWS.
  // ONDEMAND_K8S_GCP step is when user is connecting to aqueduct cluster on GCP.
  // ONDEMAND_K8S_AZURE step is when user is connecting to aqueduct cluster on Azure.

  const handleRegularK8s = () => {
    setCurrentStep('REGULAR_K8S');
  };

  const handleOndemandK8s = () => {
    setCurrentStep('ONDEMAND_K8S');
  };

  const handlePrevious = () => {
    setCurrentStep('INITIAL');
  };

  const handleAWSClick = () => {
    setCurrentStep('ONDEMAND_K8S_AWS');
  };

  const InitialStepLayout: React.FC<IntegrationDialogProps> = ({
    editMode = false,
    onCloseDialog,
    loading,
    disabled,
  }) => {
    return (
      <Box>
        <DialogTitle sx={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <div>
            <Typography variant="h5" sx={{ color: 'black' }}>
              Connect to Kubernetes
            </Typography>
          </div>
        </DialogTitle>
        <DialogContent sx={{ marginTop: '8px' }}>
          <Button
            sx={{
              textTransform: 'none',
              marginBottom: '8px',
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
            }}
            onClick={handleRegularK8s}
          >
            <IntegrationLogo
              service={`Kubernetes`}
              activated={true}
              size="small"
            />
            <Typography
              variant="body2"
              sx={{ color: 'black', fontSize: '18px' }}
            >
              I have an existing Kubernetes cluster I&apos;d like to use
            </Typography>
          </Button>
          <Button
            sx={{
              textTransform: 'none',
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
            }}
            onClick={handleOndemandK8s}
          >
            <IntegrationLogo
              service={'Aqueduct'}
              activated={SupportedIntegrations['Aqueduct'].activated}
              size="small"
            />
            <Typography
              variant="body2"
              sx={{ color: 'black', fontSize: '18px' }}
            >
              I&apos;d like Aqueduct to create & manage a cluster for me
            </Typography>
          </Button>
        </DialogContent>
        <DialogActions>
          <Button autoFocus onClick={onCloseDialog}>
            Cancel
          </Button>
        </DialogActions>
      </Box>
    );
  };

  // We're going to need to share some more info with the dialogs, as they're not all just forms that we can
  // register anymore in the case of this layout.
  const RegularK8sStepLayout: React.FC<IntegrationDialogProps> = ({
    editMode,
    onCloseDialog,
    loading,
    disabled,
  }) => {
    const methods = useFormContext();
    const dispatch: AppDispatch = useDispatch();
    const { user } = useUser();

    return (
      <>
        <DialogHeader integrationToEdit={undefined} service={'Kubernetes'} />
        <KubernetesDialog
          editMode={false}
          onCloseDialog={onCloseDialog}
          loading={loading}
          disabled={disabled}
        />
        <DialogActionButtons
          onCloseDialog={onCloseDialog}
          loading={loading}
          disabled={disabled}
          onSubmit={async () => {
            await methods.handleSubmit((data) => {
              dispatch(
                handleConnectToNewIntegration({
                  apiKey: user.apiKey,
                  service: 'Kubernetes' as Service,
                  name: data.name,
                  config: data,
                })
              );
            })(); // Remember the last two parens to call the function!
          }}
        />
      </>
    );
  };

  const OnDemandK8sStep: React.FC<IntegrationDialogProps> = ({
    editMode,
    onCloseDialog,
    loading,
    disabled,
  }) => {
    return (
      <>
        <DialogTitle sx={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <IntegrationLogo
            service={'Aqueduct'}
            activated={SupportedIntegrations['Aqueduct'].activated}
            size="small"
          />
          <div>
            <Typography variant="h5" sx={{ color: 'black' }}>
              +
            </Typography>
          </div>
          <IntegrationLogo
            service={'Kubernetes'}
            activated={SupportedIntegrations['Kubernetes'].activated}
            size="small"
          />
          <div>
            <Typography variant="h5" sx={{ color: 'black' }}>
              Aqueduct-managed Kubernetes
            </Typography>
          </div>
        </DialogTitle>
        <DialogContent
          sx={{
            display: 'flex',
            alignItems: 'center',
            paddingLeft: '54px',
            gap: '32px',
            marginTop: '16px',
            '& button': { backgroundColor: '#F8F8F8' },
          }}
        >
          <Button onClick={handleAWSClick}>
            <IntegrationLogo
              service={'Amazon'}
              activated={SupportedIntegrations['Amazon'].activated}
              size="large"
            />
          </Button>
          <Button disabled={true}>
            <IntegrationLogo
              service={'GCP'}
              activated={SupportedIntegrations['GCP'].activated}
              size="large"
            />
          </Button>
          <Button disabled={true}>
            <IntegrationLogo
              service={'Azure'}
              activated={SupportedIntegrations['Azure'].activated}
              size="large"
            />
          </Button>
        </DialogContent>
        <DialogActions>
          <Button autoFocus onClick={handlePrevious}>
            Previous
          </Button>
          <Button autoFocus onClick={onCloseDialog}>
            Cancel
          </Button>
        </DialogActions>
      </>
    );
  };

  const OnDemandK8sAWSStep: React.FC<IntegrationDialogProps> = ({
    editMode,
    onCloseDialog,
    loading,
    disabled,
  }) => {
    return (
      <>
        <DialogTitle sx={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <IntegrationLogo
            service={'Aqueduct'}
            activated={SupportedIntegrations['Aqueduct'].activated}
            size="small"
          />
          <div>
            <Typography variant="h5" sx={{ color: 'black' }}>
              +
            </Typography>
          </div>
          <IntegrationLogo
            service={'AWS'}
            activated={SupportedIntegrations['Kubernetes'].activated}
            size="small"
          />
          <div>
            <Typography variant="h5" sx={{ color: 'black' }}>
              Aqueduct-managed Kubernetes
            </Typography>
          </div>
        </DialogTitle>
        <AWSDialog editMode={false} />
        <DialogActionButtons
          onCloseDialog={onCloseDialog}
          loading={loading}
          disabled={disabled}
        />
      </>
    );
  };

  switch (currentStep) {
    case 'INITIAL':
      return (
        <InitialStepLayout
          disabled={disabled}
          loading={loading}
          onCloseDialog={onCloseDialog}
          editMode={editMode}
        />
      );
    case 'REGULAR_K8S':
      return (
        <RegularK8sStepLayout
          disabled={disabled}
          loading={loading}
          onCloseDialog={onCloseDialog}
          editMode={editMode}
        />
      );
    case 'ONDEMAND_K8S':
      return (
        <OnDemandK8sStep
          disabled={disabled}
          loading={loading}
          onCloseDialog={onCloseDialog}
          editMode={editMode}
        />
      );
    case 'ONDEMAND_K8S_AWS':
      return (
        <OnDemandK8sAWSStep
          disabled={disabled}
          loading={loading}
          onCloseDialog={onCloseDialog}
          editMode={editMode}
        />
      );
    default:
      return (
        <InitialStepLayout
          disabled={disabled}
          loading={loading}
          onCloseDialog={onCloseDialog}
          editMode={editMode}
        />
      );
  }
};

// TODO: Conditionally validate based on current step's value
export function getOnDemandKubernetesValidationSchema() {
  // Validation schema for kubernetes dialog

  // Kubernetes validation schema:
  return Yup.object().shape({
    use_same_cluster: Yup.string(),
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

export default OnDemandKubernetesDialog;
