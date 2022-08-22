import { Checkbox, FormControlLabel } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';

import { Tab, Tabs } from '../../../components/primitives/Tabs.styles';
import {
  FileData,
  IntegrationConfig,
  S3Config,
  S3CredentialType,
} from '../../../utils/integrations';
import { readCredentialsFile } from './bigqueryDialog';
import { IntegrationFileUploadField } from './IntegrationFileUploadField';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: S3Config = {
  type: S3CredentialType.AccessKey,
  bucket: 'aqueduct',
  region: 'us-east-1',
  access_key_id: '',
  secret_access_key: '',
  config_file_path: '',
  config_file_content: '',
  config_file_profile: '',
  use_as_storage: '',
};

type Props = {
  setDialogConfig: (config: IntegrationConfig) => void;
};

export const S3Dialog: React.FC<Props> = ({ setDialogConfig }) => {
  const [bucket, setBucket] = useState<string>(null);
  const [region, setRegion] = useState<string>(null);
  const [accessKeyId, setAccessKeyId] = useState<string>(null);
  const [secretAccessKey, setSecretAccessKey] = useState<string>(null);
  const [configFilePath, setConfigFilePath] = useState<string>(null);
  const [file, setFile] = useState<FileData>(null);
  const [configFileProfile, setConfigFileProfile] = useState<string>(null);
  const [s3Type, setS3Type] = useState<S3CredentialType>(
    S3CredentialType.AccessKey
  );
  const [useAsStorage, setUseAsStorage] = useState<string>('false');

  useEffect(() => {
    const config: S3Config = {
      type: s3Type,
      bucket: bucket,
      region: region,
      access_key_id: accessKeyId,
      secret_access_key: secretAccessKey,
      config_file_path: configFilePath,
      config_file_content: file?.data ?? '',
      config_file_profile: configFileProfile,
      use_as_storage: useAsStorage,
    };
    setDialogConfig(config);
  }, [
    bucket,
    region,
    accessKeyId,
    secretAccessKey,
    configFilePath,
    file,
    configFileProfile,
    s3Type,
    useAsStorage,
  ]);

  const configProfileInput = (
    <IntegrationTextInputField
      spellCheck={false}
      required={true}
      label="AWS Profile*"
      description="The name of the profile specified in brackets in your credential file."
      placeholder={Placeholders.secret_access_key}
      onChange={(event) => setConfigFileProfile(event.target.value)}
      value={secretAccessKey}
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
        onChange={(event) => setAccessKeyId(event.target.value)}
        value={accessKeyId}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="AWS Secret Access Key*"
        description="The secret access key of your AWS account."
        placeholder={Placeholders.secret_access_key}
        onChange={(event) => setSecretAccessKey(event.target.value)}
        value={secretAccessKey}
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
        description={'The absolute path to the credentials file'}
        placeholder={Placeholders.access_key_id}
        onChange={(event) => setConfigFilePath(event.target.value)}
        value={accessKeyId}
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
      {/* add these message once integration edit is ready:
        Once connected, you would need to re-upload the file to update the credentials.
      */}
      <IntegrationFileUploadField
        label={'AWS Credentials File*'}
        description={'Upload your credentials file here.'}
        required={true}
        file={file}
        placeholder={''}
        onFiles={(files) => {
          const file = files[0];
          readCredentialsFile(file, setFile);
        }}
        displayFile={null}
        onReset={(_) => {
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
        onChange={(event) => setBucket(event.target.value)}
        value={bucket}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Region*"
        description="The region the S3 bucket belongs to."
        placeholder={Placeholders.region}
        onChange={(event) => setRegion(event.target.value)}
        value={region}
      />

      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}>
        <Tabs value={s3Type} onChange={(_, value) => setS3Type(value)}>
          <Tab value={S3CredentialType.AccessKey} label="Enter Access Keys" />
          <Tab
            value={S3CredentialType.ConfigFilePath}
            label="Specify Path to Credentials"
          />
          <Tab
            value={S3CredentialType.ConfigFileContent}
            label="Upload Credentials File"
          />
        </Tabs>
      </Box>
      {s3Type === S3CredentialType.AccessKey && accessKeyTab}
      {s3Type === S3CredentialType.ConfigFilePath && configPathTab}
      {s3Type === S3CredentialType.ConfigFileContent && configUploadTab}

      <FormControlLabel
        label="Use this integration for Aqueduct metadata storage."
        control={<Checkbox checked={useAsStorage === 'true'} onChange={(event) => setUseAsStorage(event.target.checked ? 'true' : 'false')} />}
      />
    </Box>
  );
};

export function isS3ConfigComplete(config: S3Config): boolean {
  if (!config.bucket) {
    return false;
  }

  if (config.type === S3CredentialType.AccessKey) {
    return !!config.access_key_id && !!config.secret_access_key;
  }

  if (config.type === S3CredentialType.ConfigFilePath) {
    return !!config.config_file_profile && !!config.config_file_path;
  }

  if (config.type === S3CredentialType.ConfigFileContent) {
    return !!config.config_file_profile && !!config.config_file_content;
  }

  return false;
}
