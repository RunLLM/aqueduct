import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useState } from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import {
  AWSConfig,
  DynamicEngineType,
  DynamicK8sConfig,
  IntegrationDialogProps,
} from '../../../utils/integrations';
import { AWSCredentialType } from '../../../utils/shared';
import { Tab, Tabs } from '../../primitives/Tabs.styles';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: AWSConfig = {
  type: AWSCredentialType.AccessKey,
  region: 'us-east-2',
  access_key_id: '',
  secret_access_key: '',
  config_file_path: '',
  config_file_profile: '',
  k8s_serialized: '',
};

const K8sPlaceholders: DynamicK8sConfig = {
  keepalive: '1200',
  cpu_node_type: 't3.xlarge',
  gpu_node_type: 'p2.xlarge',
  min_cpu_node: '1',
  max_cpu_node: '1',
  min_gpu_node: '0',
  max_gpu_node: '1',
};

export const AWSDialog: React.FC<IntegrationDialogProps> = () => {
  const { register, getValues, setValue } = useFormContext();

  // Need state variable to change tabs, as the formContext doesn't change as readily.
  const [currentTab, setCurrentTab] = useState(AWSCredentialType.AccessKey);
  const [engineTypeTab, setEngineTypeTab] = useState(DynamicEngineType.K8s);

  register('engineType', { value: DynamicEngineType.K8s });
  register('type', { value: AWSCredentialType.AccessKey });
  register('k8s_serialized', { value: '{}' });

  const k8s_serialized = getValues('k8s_serialized');
  const k8sConfigs = JSON.parse(k8s_serialized ?? '{}') as {
    [key: string]: string;
  };

  const configProfileInput = (
    <IntegrationTextInputField
      name="config_file_profile"
      spellCheck={false}
      required={true}
      label="AWS Profile*"
      description="The name of the profile specified in brackets in your credential file."
      placeholder={Placeholders.config_file_profile}
      onChange={(event) => setValue('config_file_profile', event.target.value)}
    />
  );

  const accessKeyTab = (
    <Box>
      <Typography variant="body2" color="gray.700">
        Manually enter your AWS credentials.
      </Typography>
      <IntegrationTextInputField
        name="access_key_id"
        spellCheck={false}
        required={true}
        label="AWS Access Key ID*"
        description="The access key ID of your AWS account."
        placeholder={Placeholders.access_key_id}
        onChange={(event) => setValue('access_key_id', event.target.value)}
      />

      <IntegrationTextInputField
        name="secret_access_key"
        spellCheck={false}
        required={true}
        label="AWS Secret Access Key*"
        description="The secret access key of your AWS account."
        placeholder={Placeholders.secret_access_key}
        onChange={(event) => setValue('secret_access_key', event.target.value)}
      />

      <IntegrationTextInputField
        name="region"
        spellCheck={false}
        required={true}
        label="AWS Region*"
        description="The region of your AWS account."
        placeholder={Placeholders.region}
        onChange={(event) => setValue('region', event.target.value)}
      />
    </Box>
  );

  const configPathTab = (
    <Box>
      <Typography variant="body2" color="gray.700">
        Specify the path to your AWS credentials <strong>on the machine</strong>{' '}
        where you are running the Aqueduct server. Typically, this is in{' '}
        <code>~/.aws/credentials</code>, or <code>~/.aws/config</code> for SSO.
        You also need to specify the profile name you would like to use for the
        credentials file. Once connected, any updates to the file content will
        automatically apply to this integration.
      </Typography>
      <IntegrationTextInputField
        name="config_file_path"
        spellCheck={false}
        required={true}
        label="AWS Credentials File Path*"
        description={'The path to the credentials file'}
        placeholder={Placeholders.config_file_path}
        onChange={(event) => setValue('config_file_path', event.target.value)}
      />

      {configProfileInput}
    </Box>
  );

  const k8sConfigTab = (
    <Box>
      <Typography variant="body2" color="gray.700">
        Optionally configure on-demand Kubernetes cluster parameters.
      </Typography>
      <IntegrationTextInputField
        name="keepalive"
        spellCheck={false}
        required={false}
        label="Keepalive period"
        description="How long (in seconds) does the cluster need to remain idle before it is deleted."
        placeholder={K8sPlaceholders.keepalive}
        onChange={(event) => {
          setValue('keepalive', event.target.value);
          k8sConfigs['keepalive'] = event.target.value;
          setValue('k8s_serialized', JSON.stringify(k8sConfigs));
        }}
      />
      <IntegrationTextInputField
        name="cpu_node_type"
        spellCheck={false}
        required={false}
        label="CPU node type"
        description="The EC2 instance type of the CPU node group."
        placeholder={K8sPlaceholders.cpu_node_type}
        onChange={(event) => {
          setValue('cpu_node_type', event.target.value);
          k8sConfigs['cpu_node_type'] = event.target.value;
          setValue('k8s_serialized', JSON.stringify(k8sConfigs));
        }}
      />

      <IntegrationTextInputField
        name="gpu_node_type"
        spellCheck={false}
        required={false}
        label="GPU node type"
        description="The EC2 instance type of the GPU node group."
        placeholder={K8sPlaceholders.gpu_node_type}
        onChange={(event) => {
          setValue('gpu_node_type', event.target.value);
          k8sConfigs['gpu_node_type'] = event.target.value;
          setValue('k8s_serialized', JSON.stringify(k8sConfigs));
        }}
      />

      <IntegrationTextInputField
        name="min_cpu_node"
        spellCheck={false}
        required={false}
        label="Min CPU node"
        description="Minimum number of nodes in the CPU node group."
        placeholder={K8sPlaceholders.min_cpu_node}
        onChange={(event) => {
          setValue('min_cpu_node', event.target.value);
          k8sConfigs['min_cpu_node'] = event.target.value;
          setValue('k8s_serialized', JSON.stringify(k8sConfigs));
        }}
      />

      <IntegrationTextInputField
        name="max_cpu_node"
        spellCheck={false}
        required={false}
        label="Max CPU node"
        description="Maximum number of nodes in the CPU node group."
        placeholder={K8sPlaceholders.max_cpu_node}
        onChange={(event) => {
          setValue('max_cpu_node', event.target.value);
          k8sConfigs['max_cpu_node'] = event.target.value;
          setValue('k8s_serialized', JSON.stringify(k8sConfigs));
        }}
      />

      <IntegrationTextInputField
        name="min_gpu_node"
        spellCheck={false}
        required={false}
        label="Min GPU node"
        description="Minimum number of nodes in the GPU node group."
        placeholder={K8sPlaceholders.min_gpu_node}
        onChange={(event) => {
          setValue('min_gpu_node', event.target.value);
          k8sConfigs['min_gpu_node'] = event.target.value;
          setValue('k8s_serialized', JSON.stringify(k8sConfigs));
        }}
      />

      <IntegrationTextInputField
        name="max_gpu_node"
        spellCheck={false}
        required={false}
        label="Max GPU node"
        description="Maximum number of nodes in the GPU node group."
        placeholder={K8sPlaceholders.max_gpu_node}
        onChange={(event) => {
          setValue('max_gpu_node', event.target.value);
          k8sConfigs['max_gpu_node'] = event.target.value;
          setValue('k8s_serialized', JSON.stringify(k8sConfigs));
        }}
      />
    </Box>
  );

  return (
    <Box sx={{ mt: 2 }}>
      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}>
        <Tabs
          value={currentTab}
          onChange={(_, value) => {
            setValue('type', value);
            setCurrentTab(value);
          }}
        >
          <Tab value={AWSCredentialType.AccessKey} label="Enter Access Keys" />
          <Tab
            value={AWSCredentialType.ConfigFilePath}
            label="Specify Path to Credentials"
          />
        </Tabs>
      </Box>
      {currentTab === AWSCredentialType.AccessKey && accessKeyTab}
      {currentTab === AWSCredentialType.ConfigFilePath && configPathTab}
      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}>
        <Tabs
          value={engineTypeTab}
          onChange={(_, value) => {
            setEngineTypeTab(value);
            setValue('engine_type', value);
          }}
        >
          <Tab
            value={DynamicEngineType.K8s}
            label="On-demand Kubernetes Cluster Config"
          />
        </Tabs>
      </Box>
      {engineTypeTab === DynamicEngineType.K8s && k8sConfigTab}
    </Box>
  );
};

export function getAWSValidationSchema() {
  return Yup.object().shape({
    type: Yup.string().required('Please select a credential type'),
    access_key_id: Yup.string().when('type', {
      is: 'access_key',
      then: Yup.string().required('Please enter an access key id'),
    }),
    secret_access_key: Yup.string().when('type', {
      is: 'access_key',
      then: Yup.string().required('Please enter a secret access key'),
    }),
    region: Yup.string().when('type', {
      is: 'access_key',
      then: Yup.string().required('Please enter a region'),
    }),
    config_file_profile: Yup.string().when('type', {
      is: 'config_file_path',
      then: Yup.string().required('Please enter a config file profile'),
    }),
    config_file_path: Yup.string().when('type', {
      is: 'config_file_path',
      then: Yup.string().required('Please enter a profile path'),
    }),
  });
}
