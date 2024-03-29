import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useState } from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import {
  AthenaConfig,
  FileData,
  ResourceDialogProps,
} from '../../../utils/resources';
import { Tab, Tabs } from '../../primitives/Tabs.styles';
import { readCredentialsFile } from './bigqueryDialog';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { ResourceFileUploadField } from './ResourceFileUploadField';
import { ResourceTextInputField } from './ResourceTextInputField';
import { requiredAtCreate } from './schema';

enum AWSCredentialType {
  AccessKey = 'access_key',
  ConfigFilePath = 'config_file_path',
  ConfigFileContent = 'config_file_content',
}

const Placeholders: AthenaConfig = {
  type: AWSCredentialType.AccessKey,
  access_key_id: '',
  secret_access_key: '',
  region: 'us-east-1',
  config_file_path: '',
  config_file_content: '',
  config_file_profile: '',
  database: '',
  output_location: 's3://bucket/path/to/folder/',
};

export const AthenaDialog: React.FC<ResourceDialogProps<AthenaConfig>> = ({
  resourceToEdit,
}) => {
  const initialAccessKeyType = resourceToEdit?.config_file_path
    ? AWSCredentialType.ConfigFilePath
    : resourceToEdit?.config_file_content
    ? AWSCredentialType.ConfigFileContent
    : AWSCredentialType.AccessKey;
  const [fileData, setFileData] = useState<FileData | null>(null);
  // Need state variable to change tabs, as the formContext doesn't change as readily.
  const [currentTab, setCurrentTab] = useState(initialAccessKeyType);
  const { setValue, register } = useFormContext();
  if (resourceToEdit) {
    Object.entries(resourceToEdit).forEach(([k, v]) => {
      register(k, { value: v });
    });
  }

  const editMode = !!resourceToEdit;

  register('type', { value: initialAccessKeyType, required: true });

  const setFile = (fileData: FileData | null) => {
    // Update the react-hook-form value
    setValue('config_file_content', fileData?.data);
    // Set state to trigger re-render of file upload field.
    setFileData(fileData);
  };

  const configProfileInput = (
    <ResourceTextInputField
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
      <ResourceTextInputField
        name="access_key_id"
        spellCheck={false}
        required={true}
        label="AWS Access Key ID*"
        description="The access key ID of your AWS account."
        placeholder={Placeholders.access_key_id}
        onChange={(event) => setValue('access_key_id', event.target.value)}
      />

      <ResourceTextInputField
        name="secret_access_key"
        spellCheck={false}
        required={true}
        label="AWS Secret Access Key*"
        description="The secret access key of your AWS account."
        placeholder={Placeholders.secret_access_key}
        onChange={(event) => setValue('secret_access_key', event.target.value)}
      />

      <ResourceTextInputField
        name="region"
        spellCheck={false}
        required={true}
        label="Region*"
        description="The region the Athena database belongs to."
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
        automatically apply to this resource.
      </Typography>
      <ResourceTextInputField
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
      {/* TODO: add these message once resource edit is ready:
        Once connected, you would need to re-upload the file to update the credentials.
      */}
      <ResourceFileUploadField
        name="config_file_content"
        label={'AWS Credentials File*'}
        description={'Upload your credentials file here.'}
        required={true}
        file={fileData}
        placeholder={''}
        onFiles={(files) => {
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
      <ResourceTextInputField
        name="database"
        spellCheck={false}
        required={true}
        label="Database*"
        description="The name of the Athena database."
        placeholder={Placeholders.database}
        onChange={(event) => setValue('database', event.target.value)}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <ResourceTextInputField
        name="output_location"
        spellCheck={false}
        required={true}
        label="S3 Output Location*"
        description="The S3 path where Athena query results are written. If the path does not exist 
        in advance, Aqueduct attempts to create it. Data written to this location is garbage collected
        after each query."
        placeholder={Placeholders.output_location}
        onChange={(event) => setValue('output_location', event.target.value)}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      {/* TODO: Share tabs code with the aws and s3 dialog components. */}
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
          <Tab
            value={AWSCredentialType.ConfigFileContent}
            label="Upload Credentials File"
          />
        </Tabs>
      </Box>
      {currentTab === AWSCredentialType.AccessKey && accessKeyTab}
      {currentTab === AWSCredentialType.ConfigFilePath && configPathTab}
      {currentTab === AWSCredentialType.ConfigFileContent && configUploadTab}
    </Box>
  );
};

// Required fields are (baseFields):
// - database
// - output_location

// When using access key, also need:
// - access_key_id
// - secret_access_key
// - region

// When using credentials file, also need:
// - file path and file content
// - config_file_profile
export function getAthenaValidationSchema(editMode: boolean) {
  return Yup.object().shape({
    type: Yup.string().required('Please select a credential type'),
    database: Yup.string().required('Please enter a database name'),
    output_location: Yup.string().required('Please enter an output location'),
    access_key_id: Yup.string().when('type', {
      is: 'access_key',
      then: requiredAtCreate(
        Yup.string(),
        editMode,
        'Please enter an access key id'
      ),
    }),
    secret_access_key: Yup.string().when('type', {
      is: 'access_key',
      then: requiredAtCreate(
        Yup.string(),
        editMode,
        'Please enter a secret access key'
      ),
    }),
    region: Yup.string().when('type', {
      is: 'access_key',
      then: Yup.string().required('Please enter a region'),
    }),
    config_file_profile: Yup.string().when('type', {
      is: 'config_file_path' || 'config_file_content',
      then: Yup.string().required('Please enter a config file profile'),
    }),
    config_file_path: Yup.string().when('type', {
      is: 'config_file_path',
      then: requiredAtCreate(Yup.string(), editMode, 'Please enter a config'),
    }),
    config_file_content: Yup.string().when('type', {
      is: 'config_file_content',
      then: requiredAtCreate(
        Yup.string(),
        editMode,
        'Please upload a config file.'
      ),
    }),
  });
}
