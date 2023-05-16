import { Checkbox, FormControlLabel } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useState } from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import {
  FileData,
  IntegrationDialogProps,
  S3Config,
} from '../../../utils/integrations';
import { AWSCredentialType } from '../../../utils/shared';
import { Tab, Tabs } from '../../primitives/Tabs.styles';
import { readCredentialsFile } from './bigqueryDialog';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { IntegrationFileUploadField } from './IntegrationFileUploadField';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: S3Config = {
  type: AWSCredentialType.AccessKey,
  bucket: 'aqueduct',
  region: 'us-east-1',
  root_dir: 'path/to/root/',
  access_key_id: '',
  secret_access_key: '',
  config_file_path: '',
  config_file_content: '',
  config_file_profile: '',
  use_as_storage: '',
};

interface S3DialogProps extends IntegrationDialogProps {
  setMigrateStorage: (value: boolean) => void;
}

export const S3Dialog: React.FC<S3DialogProps> = ({
  editMode = false,
  setMigrateStorage,
}) => {
  const [fileData, setFileData] = useState<FileData | null>(null);

  const { register, setValue } = useFormContext();
  register('use_as_storage');
  const [useAsMetadataStorage, setUseAsMetadataStorage] =
    useState<string>('false');

  const [currentTab, setCurrentTab] = useState<AWSCredentialType>(
    AWSCredentialType['AccessKey']
  );
  register('type', { value: currentTab, required: true });

  const setFile = (fileData: FileData | null) => {
    // Update the react-hook-form value
    setValue('config_file_content', fileData);
    // Set state to trigger re-render of file upload field.
    setFileData(fileData);
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

  const configUploadTab = (
    <Box>
      <Typography variant="body2" color="gray.700">
        Upload your AWS credentials file. Typically, this is in{' '}
        <code>~/.aws/credentials</code>. You also need to specify the profile
        name you would like to use for the credentials file. If you are using an
        SSO profile, you should use <code>SPECIFY PATH TO CREDENTIALS</code>{' '}
        instead.
      </Typography>
      {/* TODO: add these message once integration edit is ready:
        Once connected, you would need to re-upload the file to update the credentials.
      */}
      <IntegrationFileUploadField
        name="config_file_content"
        label={'AWS Credentials File*'}
        description={'Upload your credentials file here.'}
        required={true}
        file={fileData}
        placeholder={''}
        onFiles={(files: FileList): void => {
          const file = files[0];
          readCredentialsFile(file, setFile);
        }}
        displayFile={null}
        onReset={() => {
          setFile(null);
        }}
      />

      {configProfileInput}
    </Box>
  );

  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        name="bucket"
        spellCheck={false}
        required={true}
        label="Bucket*"
        description="The name of the S3 bucket."
        placeholder={Placeholders.bucket}
        onChange={(event) => setValue('bucket', event.target.value)}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        name="region"
        spellCheck={false}
        required={true}
        label="Region*"
        description="The region the S3 bucket belongs to."
        placeholder={Placeholders.region}
        onChange={(event) => setValue('region', event.target.value)}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        name="root_dir"
        spellCheck={false}
        required={false}
        label="Directory"
        description="Only applicable when also setting this integration to be the artifact store. This is an optional path to an existing directory in the bucket, to be used as the root of the artifact store. Defaults to the root of the bucket."
        placeholder={Placeholders.root_dir}
        onChange={(event) => setValue('root_dir', event.target.value)}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}>
        <Tabs
          value={currentTab}
          onChange={(_, value) => {
            setValue('type', value);
            // reset config_file_profile when changing tabs.
            setValue('config_file_profile', '', { shouldDirty: false, shouldTouch: false });
            setCurrentTab(value);
          }}
        >
          <Tab value={AWSCredentialType.AccessKey} label="Enter Access Keys" />
          <Tab
            value={AWSCredentialType.ConfigFilePath}
            label="Specify Path to Credentials"
          />
          <Tab
            value={AWSCredentialType.ConfigFileContent}
            label="Upload Credentials File"
          />
        </Tabs>
      </Box>
      {currentTab === AWSCredentialType.AccessKey && accessKeyTab}
      {currentTab === AWSCredentialType.ConfigFilePath && configPathTab}
      {currentTab === AWSCredentialType.ConfigFileContent && configUploadTab}

      <FormControlLabel
        label="Use this integration for Aqueduct metadata storage."
        control={
          <Checkbox
            checked={useAsMetadataStorage === 'true'}
            onChange={(event) => {
              const useAsMetadataStorageChecked = event.target.checked
                ? 'true'
                : 'false';
              // Update the react-hook-form value
              setValue('use_as_storage', useAsMetadataStorageChecked);
              // Set state so that we can trigger re-render
              setUseAsMetadataStorage(useAsMetadataStorageChecked);
              // Call MigrateStorage callback to show banner
              setMigrateStorage(event.target.checked);
            }}
            disabled={editMode}
          />
        }
      />
    </Box>
  );
};

export function getS3ValidationSchema() {
  return Yup.object().shape({
    type: Yup.string().required('Please select a credential type'),
    bucket: Yup.string().required('Please enter a bucket name'),
    region: Yup.string().required('Please enter a region'),
    access_key_id: Yup.string().when('type', {
      is: 'access_key',
      then: Yup.string().required('Please enter an access key id'),
      otherwise: null,
    }),
    secret_access_key: Yup.string().when('type', {
      is: 'access_key',
      then: Yup.string().required('Please enter a secret access key'),
      otherwise: null,
    }),
    config_file_path: Yup.string().when('type', {
      is: 'config_file_path',
      then: Yup.string().required('Please enter a profile path'),
      otherwise: null,
    }),
    config_file_profile: Yup.string().when('type', {
      is: (value) => value === 'config_file_path' || value === 'config_file_content',
      then: Yup.string().required('Please enter a config file profile'),
      otherwise: null,
    }),
    config_file_content: Yup.string().when('type', {
      is: (value) => value === 'config_file_content',
      then: Yup.string()
      .transform((value) => {
        // Depending on if dragged and dropped or uploaded via file picker, we can get two different things.
        if (typeof value === 'object') {
          return value.data;
        } else if (typeof value === 'string') {
          const parsed = JSON.parse(value);
          console.log('parsed: ', parsed);
          return parsed.data;
        }

        return value;
      })
      .required('Please upload a credentials file'),
      otherwise: null,
    }),
  });
}
