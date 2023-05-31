import Box from '@mui/material/Box';
import Link from '@mui/material/Link';
import Typography from '@mui/material/Typography';
import React, { useState } from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import {
  DynamicEngineType,
  FileData,
  GCPConfig,
  OndemandGKEConfig,
  ResourceDialogProps,
} from '../../../utils/resources';
import { Tab, Tabs } from '../../primitives/Tabs.styles';
import { ResourceFileUploadField } from './ResourceFileUploadField';
import { ResourceTextInputField } from './ResourceTextInputField';

const Placeholders: OndemandGKEConfig = {
  gcp_config_serialized: '',
  keepalive: '1200',
  cpu_node_type: 'n1-standard-4',
  gpu_node_type: 'g2-standard-4',
  min_cpu_node: '1',
  max_cpu_node: '1',
  min_gpu_node: '0',
  max_gpu_node: '1',
};

const GCPPlaceholders: GCPConfig = {
  region: 'us-central1',
  zone: 'us-central1-a',
  service_account_key: '',
};

export const GCPDialog: React.FC<
  ResourceDialogProps<OndemandGKEConfig>
> = () => {
  const { register, getValues, setValue } = useFormContext();

  const [fileData, setFileData] = useState<FileData | null>(null);

  const setFile = (fileData: FileData | null) => {
    setValue('service_account_key', fileData?.data);
    setFileData(fileData);
  };

  // Need state variable to change tabs, as the formContext doesn't change as readily.
  const [engineTypeTab, setEngineTypeTab] = useState(DynamicEngineType.K8s);

  register('engineType', { value: DynamicEngineType.K8s });
  register('gcp_config_serialized', { value: '{}' });

  const gcp_config_serialized = getValues('gcp_config_serialized');
  const gcpConfigs = JSON.parse(gcp_config_serialized ?? '{}') as {
    [key: string]: string;
  };

  const k8sConfigTab = (
    <Box>
      <Typography variant="body2" color="gray.700">
        Optionally configure on-demand Kubernetes cluster parameters.
      </Typography>
      <ResourceTextInputField
        name="keepalive"
        spellCheck={false}
        required={false}
        label="Keepalive period"
        description="How long (in seconds) does the cluster need to remain idle before it is deleted."
        placeholder={Placeholders.keepalive}
        onChange={(event) => {
          setValue('keepalive', event.target.value);
        }}
      />
      <ResourceTextInputField
        name="cpu_node_type"
        spellCheck={false}
        required={false}
        label="CPU node type"
        description="The EC2 instance type of the CPU node group."
        placeholder={Placeholders.cpu_node_type}
        onChange={(event) => {
          setValue('cpu_node_type', event.target.value);
        }}
      />

      <ResourceTextInputField
        name="gpu_node_type"
        spellCheck={false}
        required={false}
        label="GPU node type"
        description="The EC2 instance type of the GPU node group."
        placeholder={Placeholders.gpu_node_type}
        onChange={(event) => {
          setValue('gpu_node_type', event.target.value);
        }}
      />

      <ResourceTextInputField
        name="min_cpu_node"
        spellCheck={false}
        required={false}
        label="Min CPU node"
        description="Minimum number of nodes in the CPU node group."
        placeholder={Placeholders.min_cpu_node}
        onChange={(event) => {
          setValue('min_cpu_node', event.target.value);
        }}
      />

      <ResourceTextInputField
        name="max_cpu_node"
        spellCheck={false}
        required={false}
        label="Max CPU node"
        description="Maximum number of nodes in the CPU node group."
        placeholder={Placeholders.max_cpu_node}
        onChange={(event) => {
          setValue('max_cpu_node', event.target.value);
        }}
      />

      <ResourceTextInputField
        name="min_gpu_node"
        spellCheck={false}
        required={false}
        label="Min GPU node"
        description="Minimum number of nodes in the GPU node group."
        placeholder={Placeholders.min_gpu_node}
        onChange={(event) => {
          setValue('min_gpu_node', event.target.value);
        }}
      />

      <ResourceTextInputField
        name="max_gpu_node"
        spellCheck={false}
        required={false}
        label="Max GPU node"
        description="Maximum number of nodes in the GPU node group."
        placeholder={Placeholders.max_gpu_node}
        onChange={(event) => {
          setValue('max_gpu_node', event.target.value);
        }}
      />
    </Box>
  );

  const fileUploadDescription = (
    <>
      <>Follow the instructions </>
      <Link
        sx={{ fontSize: 'inherit' }}
        target="_blank"
        href="https://cloud.google.com/iam/docs/service-accounts-create"
      >
        here
      </Link>
      <> to get your service account key file.</>
    </>
  );

  return (
    <Box sx={{ mt: 2 }}>
      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}>
        <ResourceTextInputField
          name="region"
          spellCheck={false}
          required={true}
          label="Region*"
          description="The GCP region in which the cluster will be created."
          placeholder={GCPPlaceholders.region}
          onChange={(event) => {
            setValue('region', event.target.value);
            gcpConfigs['region'] = event.target.value;
            setValue('gcp_config_serialized', JSON.stringify(gcpConfigs));
          }}
        />

        <ResourceTextInputField
          name="zone"
          spellCheck={false}
          required={true}
          label="Zone*"
          description="The GCP region's zone in which the cluster will be created."
          placeholder={GCPPlaceholders.zone}
          onChange={(event) => {
            setValue('zone', event.target.value);
            gcpConfigs['zone'] = event.target.value;
            setValue('gcp_config_serialized', JSON.stringify(gcpConfigs));
          }}
        />

        <ResourceFileUploadField
          name="service_account_key"
          label={'Service Account Key File*'}
          description={fileUploadDescription}
          required={true}
          file={fileData}
          placeholder={'Upload your service account key file.'}
          onFiles={(files) => {
            const file = files[0];
            readCredentialsFile(file, (fileData) => {
              // set the fileData state
              setFile(fileData);
              // set the service_account_key in the gcpConfigs
              gcpConfigs['service_account_key'] = fileData?.data || '';
              setValue('gcp_config_serialized', JSON.stringify(gcpConfigs));
            });
          }}
          displayFile={null}
          onReset={() => {
            setFile(null);
          }}
        />
      </Box>
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

export function readCredentialsFile(
  file: File,
  callback: (credentials: FileData) => void
): void {
  const reader = new FileReader();
  reader.onloadend = function (event) {
    const content = event.target.result as string;
    callback({ name: file.name, data: content });
  };
  reader.readAsText(file);
}

export function getGCPValidationSchema() {
  return Yup.object().shape({
    name: Yup.string().required('Please enter a name'),
    region: Yup.string().required('Please enter a region'),
    zone: Yup.string().required('Please enter a zone'),
    service_account_key: Yup.string()
      .transform((value) => {
        if (!value?.data) {
          return null;
        }
        return value.data;
      })
      .required('Please upload a service account key file'),
  });
}
