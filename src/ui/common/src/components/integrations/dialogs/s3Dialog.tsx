import { Checkbox, FormControlLabel } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';

import {
  AWSCredentialType,
  FileData,
  S3Config,
} from '../../../utils/integrations';
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

type Props = {
  onUpdateField: (field: keyof S3Config, value: string) => void;
  value?: S3Config;
  editMode: boolean;
  setMigrateStorage: (value: boolean) => void;
};

export const S3Dialog: React.FC<Props> = ({
  onUpdateField,
  value,
  editMode,
  setMigrateStorage,
}) => {
  const [fileName, setFileName] = useState<string>(null);

  const setFile = (fileData: FileData | null) => {
    setFileName(fileData?.name ?? null);
    onUpdateField('config_file_content', fileData?.data);
  };

  const fileData =
    fileName && !!value?.config_file_content
      ? {
          name: fileName,
          data: value.config_file_content,
        }
      : null;

  useEffect(() => {
    if (!value?.type) {
      onUpdateField('type', AWSCredentialType.AccessKey);
    }

    if (!value?.use_as_storage) {
      onUpdateField('use_as_storage', 'false');
    }
  }, [onUpdateField, value?.type, value?.use_as_storage]);

  const configProfileInput = (
    <IntegrationTextInputField
      spellCheck={false}
      required={true}
      label="AWS Profile*"
      description="The name of the profile specified in brackets in your credential file."
      placeholder={Placeholders.config_file_profile}
      onChange={(event) =>
        onUpdateField('config_file_profile', event.target.value)
      }
      value={value?.config_file_profile ?? ''}
    />
  );

  const accessKeyTab = (
    <Box>
      <Typography variant="body2" color="gray.700">
        Manually enter your AWS credentials.
      </Typography>
      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="AWS Access Key ID*"
        description="The access key ID of your AWS account."
        placeholder={Placeholders.access_key_id}
        onChange={(event) => onUpdateField('access_key_id', event.target.value)}
        value={value?.access_key_id ?? ''}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="AWS Secret Access Key*"
        description="The secret access key of your AWS account."
        placeholder={Placeholders.secret_access_key}
        onChange={(event) =>
          onUpdateField('secret_access_key', event.target.value)
        }
        value={value?.secret_access_key ?? ''}
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
        spellCheck={false}
        required={true}
        label="AWS Credentials File Path*"
        description={'The path to the credentials file'}
        placeholder={Placeholders.config_file_path}
        onChange={(event) =>
          onUpdateField('config_file_path', event.target.value)
        }
        value={value?.config_file_path ?? ''}
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
      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Bucket*"
        description="The name of the S3 bucket."
        placeholder={Placeholders.bucket}
        onChange={(event) => onUpdateField('bucket', event.target.value)}
        value={value?.bucket ?? ''}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Region*"
        description="The region the S3 bucket belongs to."
        placeholder={Placeholders.region}
        onChange={(event) => onUpdateField('region', event.target.value)}
        value={value?.region ?? ''}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={false}
        label="Directory"
        description="Only applicable when also setting this integration to be the artifact store. This is an optional path to an existing directory in the bucket, to be used as the root of the artifact store. Defaults to the root of the bucket."
        placeholder={Placeholders.root_dir}
        onChange={(event) => onUpdateField('root_dir', event.target.value)}
        value={value?.root_dir ?? ''}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}>
        <Tabs
          value={value?.type ?? 'access_key'}
          onChange={(_, value) => onUpdateField('type', value)}
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
      {value?.type === AWSCredentialType.AccessKey && accessKeyTab}
      {value?.type === AWSCredentialType.ConfigFilePath && configPathTab}
      {value?.type === AWSCredentialType.ConfigFileContent && configUploadTab}

      <FormControlLabel
        label="Use this integration for Aqueduct metadata storage."
        control={
          <Checkbox
            checked={value?.use_as_storage === 'true'}
            onChange={(event) => {
              onUpdateField(
                'use_as_storage',
                event.target.checked ? 'true' : 'false'
              );
              setMigrateStorage(event.target.checked);
            }}
            disabled={editMode}
          />
        }
      />
    </Box>
  );
};

export function isS3ConfigComplete(config: S3Config): boolean {
  if (!config.bucket || !config.region) {
    return false;
  }

  if (config.type === AWSCredentialType.AccessKey) {
    return !!config.access_key_id && !!config.secret_access_key;
  }

  if (config.type === AWSCredentialType.ConfigFilePath) {
    return !!config.config_file_profile && !!config.config_file_path;
  }

  if (config.type === AWSCredentialType.ConfigFileContent) {
    return !!config.config_file_profile && !!config.config_file_content;
  }

  return false;
}
