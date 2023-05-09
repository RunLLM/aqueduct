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
import IntegrationLogo from '../logo';
import { AWSDialog } from './awsDialog';
import { DialogActionButtons, DialogHeader } from './dialog';
import { IntegrationTextInputField } from './IntegrationTextInputField';
import { KubernetesDialog } from './kubernetesDialog';

const K8S_TYPES = {
  // INITIAL step is when user is choosing to connect to their own or aqueduct cluster.
  INITIAL: 'INITIAL',
  // REGULAR_K8S step is when user is connecting to their own cluster.
  REGULAR_K8S: 'REGULAR_K8S',
  // ONDEMAND_K8S step is when user is connecting to aqueduct cluster.
  ONDEMAND_K8S: 'ONDEMAND_K8S',
  // ONDEMAND_K8S_AWS step is when user is connecting to aqueduct cluster on AWS.
  ONDEMAND_K8S_AWS: 'ONDEMAND_K8S_AWS',
  // Coming soon ...
  // ONDEMAND_K8S_GCP step is when user is connecting to aqueduct cluster on GCP
  ONDEMAND_K8S_GCP: 'ONDEMAND_K8S_GCP',
  // ONDEMAND_K8S_AZURE step is when user is connecting to aqueduct cluster on Azure.
  ONDEMAND_K8S_AZURE: 'ONDEMAND_K8S_AZURE',
};

export const OnDemandKubernetesDialog: React.FC<IntegrationDialogProps> = ({
  user,
  editMode = false,
  disabled,
  loading,
  onCloseDialog,
}) => {
  const { register, setValue } = useFormContext();

  const [currentStep, setCurrentStep] = useState('INITIAL');
  register('k8s_type', { value: 'INITIAL' });

  const handleRegularK8s = () => {
    setCurrentStep(K8S_TYPES.REGULAR_K8S);
    setValue('k8s_type', K8S_TYPES.REGULAR_K8S);
  };

  const handleOndemandK8s = () => {
    setCurrentStep(K8S_TYPES.ONDEMAND_K8S);
    setValue('k8s_type', K8S_TYPES.ONDEMAND_K8S);
  };

  const handlePrevious = () => {
    setCurrentStep('INITIAL');
    setValue('k8s_type', K8S_TYPES.INITIAL);
  };

  const handleAWSClick = () => {
    setCurrentStep(K8S_TYPES.ONDEMAND_K8S_AWS);
    setValue('k8s_type', K8S_TYPES.ONDEMAND_K8S_AWS);
  };

  const InitialStepLayout: React.FC<IntegrationDialogProps> = ({
    user,
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
    user,
    editMode,
    onCloseDialog,
    loading,
    disabled,
  }) => {
    const methods = useFormContext();
    const dispatch: AppDispatch = useDispatch();

    return (
      <>
        <DialogHeader integrationToEdit={undefined} service={'Kubernetes'} />
        <IntegrationTextInputField
          name="name"
          spellCheck={false}
          required={true}
          label="Name*"
          description="Provide a unique name to refer to this integration."
          placeholder={'my_kubernetes_integration'}
          onChange={(event) => {
            methods.setValue('name', event.target.value);
          }}
          disabled={false}
        />
        <KubernetesDialog
          user={user}
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
            console.log('Submitting form');
            await methods.handleSubmit((data) => {
              console.log('handleSubmnit data: ', data);
              // Remove the name field from request body to avoid pydantic errors.
              // Name needs to be passed in as a header instead. Dunno why it's not part of the body :shrug:
              const name = data.name;
              delete data.name;
              // Remove extraneous fields if they are added when filling out the form.
              delete data.k8s_type;
              delete data.type;

              console.log('data before dispatch: ', data);

              dispatch(
                handleConnectToNewIntegration({
                  apiKey: user.apiKey,
                  service: 'Kubernetes',
                  name: name,
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
    user,
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
    user,
    editMode,
    onCloseDialog,
    loading,
    disabled,
  }) => {
    const methods = useFormContext();
    const dispatch: AppDispatch = useDispatch();

    return (
      <>
        <DialogTitle sx={{ display: 'flex', alignItems: 'center' }}>
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
        <IntegrationTextInputField
          name="name"
          spellCheck={false}
          required={true}
          label="Name*"
          description="Provide a unique name to refer to this integration."
          placeholder={'my_kubernetes_integration'}
          onChange={(event) => {
            methods.setValue('name', event.target.value);
          }}
          disabled={false}
        />
        <AWSDialog
          user={user}
          disabled={disabled}
          loading={loading}
          onCloseDialog={onCloseDialog}
          editMode={false}
        />
        <DialogActionButtons
          onCloseDialog={onCloseDialog}
          loading={loading}
          disabled={disabled}
          onSubmit={async () => {
            console.log('Submitting AWS form.');
            await methods.handleSubmit((data) => {
              console.log('handleSubmnit data: ', data);
              // Remove the name field from request body to avoid pydantic errors.
              // Name needs to be passed in as a header instead. Dunno why it's not part of the body :shrug:
              const name = data.name;
              delete data.name;
              // Remove extraneous fields if they are added when filling out the form.
              delete data.k8s_type;
              delete data.type;

              console.log('data before dispatch: ', data);

              dispatch(
                handleConnectToNewIntegration({
                  apiKey: user.apiKey,
                  service: 'Kubernetes',
                  name: name,
                  config: data,
                })
              );
            })(); // Remember the last two parens to call the function!
          }}
        />
      </>
    );
  };

  switch (currentStep) {
    case 'INITIAL':
      return (
        <InitialStepLayout
          user={user}
          disabled={disabled}
          loading={loading}
          onCloseDialog={onCloseDialog}
          editMode={editMode}
        />
      );
    case 'REGULAR_K8S':
      return (
        <RegularK8sStepLayout
          user={user}
          disabled={disabled}
          loading={loading}
          onCloseDialog={onCloseDialog}
          editMode={editMode}
        />
      );
    case 'ONDEMAND_K8S':
      return (
        <OnDemandK8sStep
          user={user}
          disabled={disabled}
          loading={loading}
          onCloseDialog={onCloseDialog}
          editMode={editMode}
        />
      );
    case 'ONDEMAND_K8S_AWS':
      return (
        <OnDemandK8sAWSStep
          user={user}
          disabled={disabled}
          loading={loading}
          onCloseDialog={onCloseDialog}
          editMode={editMode}
        />
      );
    default:
      return (
        <InitialStepLayout
          user={user}
          disabled={disabled}
          loading={loading}
          onCloseDialog={onCloseDialog}
          editMode={editMode}
        />
      );
  }
};

export function getOnDemandKubernetesValidationSchema() {
  return Yup.object().shape({
    k8s_type: Yup.string(),
    // Check the fields from the kubernetes validation schema.
    use_same_cluster: Yup.string().when('k8s_type', {
      is: K8S_TYPES.REGULAR_K8S,
      then: Yup.string().required('Please select an option'),
      otherwise: null,
    }),
    kubeconfig_path: Yup.string().when('k8s_type', {
      is: K8S_TYPES.REGULAR_K8S,
      then: Yup.string().required('Please enter a kubeconfig path'),
      otherwise: null,
    }),
    cluster_name: Yup.string().when('k8s_type', {
      is: K8S_TYPES.REGULAR_K8S,
      then: Yup.string().required('Please enter a cluster name'),
      otherwise: null,
    }),
    // Checking for the AWS fields
    type: Yup.string().when('k8s_type', {
      is: K8S_TYPES.ONDEMAND_K8S_AWS,
      then: Yup.string().required('Please select a credential type'),
      otherwise: null,
    }),
    access_key_id: Yup.string().when(['k8s_type', 'type'], {
      is: (k8s_type, type) =>
        k8s_type === K8S_TYPES.ONDEMAND_K8S_AWS && type === 'access_key',
      then: Yup.string().required('Please enter an access key id'),
      otherwise: null,
    }),
    secret_access_key: Yup.string().when(['k8s_type', 'type'], {
      is: (k8s_type, type) =>
        k8s_type === K8S_TYPES.ONDEMAND_K8S_AWS && type === 'access_key',
      then: Yup.string().required('Please enter a secret access key'),
      otherwise: null,
    }),
    region: Yup.string().when(['k8s_type', 'type'], {
      is: (k8s_type, type) =>
        k8s_type === K8S_TYPES.ONDEMAND_K8S_AWS && type === 'access_key',
      then: Yup.string().required('Please enter a region'),
      otherwise: null,
    }),
    config_file_profile: Yup.string().when(['k8s_type', 'type'], {
      is: (k8s_type, type) =>
        k8s_type === K8S_TYPES.ONDEMAND_K8S_AWS && type === 'config_file_path',
      then: Yup.string().required('Please enter a config file profile'),
      otherwise: null,
    }),
    config_file_path: Yup.string().when(['k8s_type', 'type'], {
      is: (k8s_type, type) =>
        k8s_type === K8S_TYPES.ONDEMAND_K8S_AWS && type === 'config_file_path',
      then: Yup.string().required('Please enter a profile path'),
      otherwise: null,
    }),
  });
}

export default OnDemandKubernetesDialog;
