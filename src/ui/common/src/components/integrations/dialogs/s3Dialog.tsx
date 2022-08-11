import Box from '@mui/material/Box';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';

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
  access_key_id: '',
  secret_access_key: '',
  config_file_path: '',
  config_file_content: '',
  config_file_profile: '',
};

type Props = {
  setDialogConfig: (config: IntegrationConfig) => void;
};

export const S3Dialog: React.FC<Props> = ({ setDialogConfig }) => {
  const [bucket, setBucket] = useState<string>(null);
  const [accessKeyId, setAccessKeyId] = useState<string>(null);
  const [secretAccessKey, setSecretAccessKey] = useState<string>(null);
  const [configFilePath, setConfigFilePath] = useState<string>(null);
  const [file, setFile] = useState<FileData>(null);
  const [configFileProfile, setConfigFileProfile] = useState<string>(null);
  const [s3Type, setS3Type] = useState<S3CredentialType>(
    S3CredentialType.AccessKey
  );

  useEffect(() => {
    const config: S3Config = {
      type: s3Type,
      bucket: bucket,
      access_key_id: accessKeyId,
      secret_access_key: secretAccessKey,
      config_file_path: configFilePath,
      config_file_content: file?.data ?? '',
      config_file_profile: configFileProfile,
    };
    setDialogConfig(config);
  }, [
    bucket,
    accessKeyId,
    secretAccessKey,
    configFilePath,
    file,
    configFileProfile,
    s3Type,
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
        Provide the credentials by specifying the path to your aws credentials,
        together with the name of the profile you would like to use. <br />
        The path has to be in the same machine as the one you run{' '}
        <code> aqueduct start </code> . Typically, it&apos;s{' '}
        <code>~/.aws/credentials</code>. <br />
        Once connected, any updates to the file content will automatically apply
        to this integration.
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
        Upload your aws credentials file and provide the name of the profile you
        would like to use. Typically, it&apos;s <code>~/.aws/credentials</code>.
        {/* uncomment once integration edit is ready:
      <br/>
      <br/>
        Once connected, you would need to re-upload the file to update the credentials.
      */}
      </Typography>
      <IntegrationFileUploadField
        label={'AWS Credentials File*'}
        description={'upload your credentials file here.'}
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

      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}>
        <Tabs value={s3Type} onChange={(_, value) => setS3Type(value)}>
          <Tab value={S3CredentialType.AccessKey} label="Access Keys" />
          <Tab
            value={S3CredentialType.ConfigFilePath}
            label="Credentials Path"
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
