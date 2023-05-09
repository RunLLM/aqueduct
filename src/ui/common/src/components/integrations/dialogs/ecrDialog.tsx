import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useState } from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import {
  AWSCredentialType,
  ECRConfig,
  IntegrationDialogProps,
} from '../../../utils/integrations';
import { Tab, Tabs } from '../../primitives/Tabs.styles';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: ECRConfig = {
  //type: AWSCredentialType.AccessKey,
  type: 'access_key',
  region: 'us-east-2',
  access_key_id: '',
  secret_access_key: '',
  config_file_path: '',
  config_file_profile: '',
};

export const ECRDialog: React.FC<IntegrationDialogProps> = ({
  editMode = false,
}) => {
  const { register, setValue } = useFormContext();
  const [currentTab, setCurrentTab] = useState(AWSCredentialType.AccessKey);

  register('type', { value: AWSCredentialType.AccessKey });

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
    </Box>
  );
};

// NOTE: This is the same validationschema as that of awsDialog.tsx.
// Should we consolidate the two into one? I'm not sure if we wish to support other fields in the future.
export function getECRValidationSchema() {
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
