import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle/DialogTitle';
import Typography from '@mui/material/Typography';
import React, { useState } from 'react';
import { useFormContext } from 'react-hook-form';
import { useDispatch } from 'react-redux';
import { useParams } from 'react-router-dom';
import * as Yup from 'yup';

import { useEnvironmentGetQuery } from '../../../handlers/AqueductApi';
import {
  handleConnectToNewResource,
  handleEditResource,
} from '../../../reducers/resource';
import { AppDispatch } from '../../../stores/store';
import {
  AWSConfig,
  KubernetesConfig,
  ResourceDialogProps,
} from '../../../utils/resources';
import { ResourceConfig } from '../../../utils/resources';
import SupportedResources from '../../../utils/SupportedResources';
import ResourceLogo from '../logo';
import { AWSDialog } from './awsDialog';
import { DialogActionButtons, DialogHeader } from './dialog';
import { KubernetesDialog } from './kubernetesDialog';
import { ResourceTextInputField } from './ResourceTextInputField';
import { requiredAtCreate } from './schema';

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

export const OnDemandKubernetesDialog: React.FC<
  ResourceDialogProps<ResourceConfig>
> = ({ user, resourceToEdit, disabled, loading, onCloseDialog }) => {
  const { data: environment } = useEnvironmentGetQuery({ apiKey: user.apiKey });
  const { register, setValue } = useFormContext();
  // This hack is rather hacky as we check the resource field rather than explicitly
  // pass around its service type.
  const initialStep = (resourceToEdit ?? {})['k8s_serialized']
    ? K8S_TYPES.ONDEMAND_K8S_AWS
    : (resourceToEdit ?? {})['cluster_name']
    ? K8S_TYPES.REGULAR_K8S
    : K8S_TYPES.INITIAL;

  const [currentStep, setCurrentStep] = useState(initialStep);
  register('k8s_type', { value: initialStep });

  const handleRegularK8s = () => {
    setCurrentStep(K8S_TYPES.REGULAR_K8S);
    setValue('k8s_type', K8S_TYPES.REGULAR_K8S);
  };

  const handleOndemandK8s = () => {
    setCurrentStep(K8S_TYPES.ONDEMAND_K8S);
    setValue('k8s_type', K8S_TYPES.ONDEMAND_K8S);
  };

  const handleToSelectK8sTypeStep = () => {
    setCurrentStep('INITIAL');
    setValue('k8s_type', K8S_TYPES.INITIAL);
  };

  const onSelectProvider = (provider: 'AWS' | 'GCP' | 'Azure') => {
    if (provider === 'AWS') {
      setCurrentStep(K8S_TYPES.ONDEMAND_K8S_AWS);
      setValue('k8s_type', K8S_TYPES.ONDEMAND_K8S_AWS);
      return;
    }

    if (provider === 'GCP') {
      setCurrentStep(K8S_TYPES.ONDEMAND_K8S_GCP);
      setValue('k8s_type', K8S_TYPES.ONDEMAND_K8S_GCP);
      return;
    }

    setCurrentStep(K8S_TYPES.ONDEMAND_K8S_AZURE);
    setValue('k8s_type', K8S_TYPES.ONDEMAND_K8S_AZURE);
    return;
  };

  switch (currentStep) {
    case 'INITIAL':
      return (
        <SelectK8sTypeDialog
          onCloseDialog={onCloseDialog}
          handleOnDemandK8s={handleOndemandK8s}
          handleRegularK8s={handleRegularK8s}
        />
      );
    case 'REGULAR_K8S':
      return (
        <StaticK8sDialog
          user={user}
          disabled={disabled}
          loading={loading}
          onCloseDialog={onCloseDialog}
          resourceToEdit={resourceToEdit as KubernetesConfig}
          inK8sCluster={environment?.inK8sCluster}
        />
      );
    case 'ONDEMAND_K8S':
      return (
        <SelectOnDemandCloudProviderDialog
          onCloseDialog={onCloseDialog}
          onSelectProvider={onSelectProvider}
          handleToPreviousStep={handleToSelectK8sTypeStep}
        />
      );
    case 'ONDEMAND_K8S_AWS':
      return (
        <OndemandK8sAWSDialog
          user={user}
          disabled={disabled}
          loading={loading}
          onCloseDialog={onCloseDialog}
          resourceToEdit={resourceToEdit as AWSConfig}
        />
      );
    default:
      return (
        <SelectK8sTypeDialog
          onCloseDialog={onCloseDialog}
          handleOnDemandK8s={handleOndemandK8s}
          handleRegularK8s={handleRegularK8s}
        />
      );
  }
};

type SelectK8sTypeDialogProps = {
  onCloseDialog: () => void;
  handleRegularK8s: () => void;
  handleOnDemandK8s: () => void;
};

const SelectK8sTypeDialog: React.FC<SelectK8sTypeDialogProps> = ({
  onCloseDialog,
  handleRegularK8s,
  handleOnDemandK8s,
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
          <ResourceLogo service={`Kubernetes`} activated={true} size="small" />
          <Typography variant="body2" sx={{ color: 'black', fontSize: '18px' }}>
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
          onClick={handleOnDemandK8s}
        >
          <ResourceLogo
            service={'Aqueduct'}
            activated={SupportedResources['Aqueduct'].activated}
            size="small"
          />
          <Typography variant="body2" sx={{ color: 'black', fontSize: '18px' }}>
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

const OndemandK8sAWSDialog: React.FC<ResourceDialogProps<ResourceConfig>> = ({
  user,
  resourceToEdit,
  onCloseDialog,
  loading,
  disabled,
}) => {
  const { register, setValue, handleSubmit } = useFormContext();
  const dispatch: AppDispatch = useDispatch();
  const editMode = !!resourceToEdit;
  if (resourceToEdit) {
    Object.entries(resourceToEdit).forEach(([k, v]) => {
      register(k, { value: v });
    });
  }

  // This is slightly hacky for now as we only pass around the config to edit to the dialog,
  // rather than the entire resource.
  const resourceId: string = useParams().id;

  return (
    <>
      <DialogTitle sx={{ display: 'flex', alignItems: 'center' }}>
        <ResourceLogo
          service={'Aqueduct'}
          activated={SupportedResources['Aqueduct'].activated}
          size="small"
        />
        <div>
          <Typography variant="h5" sx={{ color: 'black' }}>
            +
          </Typography>
        </div>
        <ResourceLogo
          service={'AWS'}
          activated={SupportedResources['Kubernetes'].activated}
          size="small"
        />
        <div>
          <Typography variant="h5" sx={{ color: 'black' }}>
            Aqueduct-managed Kubernetes
          </Typography>
        </div>
      </DialogTitle>
      <ResourceTextInputField
        name="name"
        spellCheck={false}
        required={true}
        label="Name*"
        description="Provide a unique name to refer to this resource."
        placeholder={'my_kubernetes_resource'}
        onChange={(event) => {
          setValue('name', event.target.value);
        }}
        disabled={false}
      />
      <AWSDialog
        user={user}
        disabled={disabled}
        loading={loading}
        onCloseDialog={onCloseDialog}
        resourceToEdit={resourceToEdit as AWSConfig}
      />
      <DialogActionButtons
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
        onSubmit={async () => {
          await handleSubmit((data) => {
            // Remove the name field from request body to avoid pydantic errors.
            // Name needs to be passed in as a header instead. Dunno why it's not part of the body :shrug:
            const name = data.name;
            const config = { ...data };
            delete config.name;
            // Remove extraneous fields if they are added when filling out the form.
            delete config.k8s_type;
            delete config.type;

            editMode
              ? dispatch(
                  handleEditResource({
                    apiKey: user.apiKey,
                    resourceId,
                    name,
                    config,
                  })
                )
              : dispatch(
                  handleConnectToNewResource({
                    apiKey: user.apiKey,
                    service: 'AWS',
                    name,
                    config,
                  })
                );
          })(); // Remember the last two parens to call the function!
        }}
      />
    </>
  );
};

type SelectOnDemandCloudProviderDialogProps = {
  onCloseDialog: () => void;
  handleToPreviousStep: () => void;
  onSelectProvider: (provider: 'GCP' | 'AWS' | 'Azure') => void;
};

const SelectOnDemandCloudProviderDialog: React.FC<
  SelectOnDemandCloudProviderDialogProps
> = ({ onCloseDialog, handleToPreviousStep, onSelectProvider }) => {
  return (
    <>
      <DialogTitle sx={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
        <ResourceLogo
          service={'Aqueduct'}
          activated={SupportedResources['Aqueduct'].activated}
          size="small"
        />
        <div>
          <Typography variant="h5" sx={{ color: 'black' }}>
            +
          </Typography>
        </div>
        <ResourceLogo
          service={'Kubernetes'}
          activated={SupportedResources['Kubernetes'].activated}
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
        <Button onClick={() => onSelectProvider('AWS')}>
          <ResourceLogo
            service={'AWS'}
            activated={SupportedResources['Amazon'].activated}
            size="large"
          />
        </Button>
        <Button disabled={true} onClick={() => onSelectProvider('GCP')}>
          <ResourceLogo
            service={'GCP'}
            activated={SupportedResources['GCP'].activated}
            size="large"
          />
        </Button>
        <Button disabled={true} onClick={() => onSelectProvider('Azure')}>
          <ResourceLogo
            service={'Azure'}
            activated={SupportedResources['Azure'].activated}
            size="large"
          />
        </Button>
      </DialogContent>
      <DialogActions>
        <Button autoFocus onClick={handleToPreviousStep}>
          Previous
        </Button>
        <Button autoFocus onClick={onCloseDialog}>
          Cancel
        </Button>
      </DialogActions>
    </>
  );
};

interface StaticK8sDialogProps extends ResourceDialogProps<KubernetesConfig> {
  inK8sCluster?: boolean;
}
// We're going to need to share some more info with the dialogs, as they're not all just forms that we can
// register anymore in the case of this layout.
const StaticK8sDialog: React.FC<StaticK8sDialogProps> = ({
  user,
  resourceToEdit,
  onCloseDialog,
  loading,
  disabled,
  inK8sCluster = false,
}) => {
  const { setValue, handleSubmit } = useFormContext();
  const dispatch: AppDispatch = useDispatch();

  return (
    <>
      <DialogHeader resourceToEdit={undefined} service={'Kubernetes'} />
      <ResourceTextInputField
        name="name"
        spellCheck={false}
        required={true}
        label="Name*"
        description="Provide a unique name to refer to this resource."
        placeholder={'my_kubernetes_resource'}
        onChange={(event) => {
          setValue('name', event.target.value);
        }}
        disabled={false}
      />
      <KubernetesDialog
        user={user}
        resourceToEdit={resourceToEdit}
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
        inK8sCluster={inK8sCluster}
      />
      <DialogActionButtons
        onCloseDialog={onCloseDialog}
        loading={loading}
        disabled={disabled}
        onSubmit={async () => {
          await handleSubmit((data) => {
            // Remove the name field from request body to avoid pydantic errors.
            // Name needs to be passed in as a header instead. Dunno why it's not part of the body :shrug:
            const name = data.name;
            const config = { ...data };
            delete config.name;
            // Remove extraneous fields if they are added when filling out the form.
            delete config.k8s_type;
            delete config.type;

            dispatch(
              handleConnectToNewResource({
                apiKey: user.apiKey,
                service: 'Kubernetes',
                name: name,
                config,
              })
            );
          })(); // Remember the last two parens to call the function!
        }}
      />
    </>
  );
};

export function getOnDemandKubernetesValidationSchema(editMode: boolean) {
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
      then: requiredAtCreate(
        Yup.string(),
        editMode,
        'Please enter an access key id'
      ),
      otherwise: null,
    }),
    secret_access_key: Yup.string().when(['k8s_type', 'type'], {
      is: (k8s_type, type) =>
        k8s_type === K8S_TYPES.ONDEMAND_K8S_AWS && type === 'access_key',
      then: requiredAtCreate(
        Yup.string(),
        editMode,
        'Please enter a secret access key'
      ),
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
      then: requiredAtCreate(
        Yup.string(),
        editMode,
        'Please enter a config file profile'
      ),
      otherwise: null,
    }),
    config_file_path: Yup.string().when(['k8s_type', 'type'], {
      is: (k8s_type, type) =>
        k8s_type === K8S_TYPES.ONDEMAND_K8S_AWS && type === 'config_file_path',
      then: requiredAtCreate(
        Yup.string(),
        editMode,
        'Please enter a profile path'
      ),
      otherwise: null,
    }),
  });
}

export default OnDemandKubernetesDialog;
