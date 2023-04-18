import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';

import {
  AthenaConfig,
  AWSCredentialType,
  FileData,
} from '../../../utils/integrations';
import { Tab, Tabs } from '../../primitives/Tabs.styles';
import { readCredentialsFile } from './bigqueryDialog';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { IntegrationFileUploadField } from './IntegrationFileUploadField';
import { IntegrationTextInputField } from './IntegrationTextInputField';

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

type Props = {
  onUpdateField: (field: keyof AthenaConfig, value: string) => void;
  value?: AthenaConfig;
  editMode: boolean;
};

export const AthenaDialog: React.FC<Props> = ({
  onUpdateField,
  value,
  editMode,
}) => {
  const [fileName, setFileName] = useState<string>(null);

  const setFile = (fileData: FileData | null) => {
    setFileName(fileData?.name ?? '');
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
  }, [onUpdateField, value?.type]);

  const configProfileInput = (
    <IntegrationTextInputField
      name="config_file_profile"
      spellCheck={false}
      required={true}
      label="AWS Profile*"
      description="The name of the profile specified in brackets in your credential file."
      placeholder={Placeholders.config_file_profile}
      onChange={(event) =>
        onUpdateField('config_file_profile', event.target.value)
      }
      //value={value?.config_file_profile ?? ''}
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
        onChange={(event) => onUpdateField('access_key_id', event.target.value)}
      />

      <IntegrationTextInputField
        name="secret_access_key"
        spellCheck={false}
        required={true}
        label="AWS Secret Access Key*"
        description="The secret access key of your AWS account."
        placeholder={Placeholders.secret_access_key}
        onChange={(event) =>
          onUpdateField('secret_access_key', event.target.value)
        }
      />

      <IntegrationTextInputField
        name="region"
        spellCheck={false}
        required={true}
        label="Region*"
        description="The region the Athena database belongs to."
        placeholder={Placeholders.region}
        onChange={(event) => onUpdateField('region', event.target.value)}
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
        onChange={(event) =>
          onUpdateField('config_file_path', event.target.value)
        }
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
        name="database"
        spellCheck={false}
        required={true}
        label="Database*"
        description="The name of the Athena database."
        placeholder={Placeholders.database}
        onChange={(event) => onUpdateField('database', event.target.value)}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        name="output_location"
        spellCheck={false}
        required={true}
        label="S3 Output Location*"
        description="The S3 path where Athena query results are written. If the path does not exist 
        in advance, Aqueduct attempts to create it. Data written to this location is garbage collected
        after each query."
        placeholder={Placeholders.output_location}
        onChange={(event) =>
          onUpdateField('output_location', event.target.value)
        }
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
    </Box>
  );
};


// TODO: Add custom Control component to render tabs and use that to update form context.

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
export function isAthenaConfigComplete(config: AthenaConfig): boolean {
  const baseFields = !!config.database && !!config.output_location;

  if (config.type === AWSCredentialType.AccessKey) {
    return (
      baseFields &&
      !!config.access_key_id &&
      !!config.secret_access_key &&
      !!config.region
    );
  }

  if (config.type === AWSCredentialType.ConfigFilePath) {
    return (
      baseFields && !!config.config_file_profile && !!config.config_file_path
    );
  }

  if (config.type === AWSCredentialType.ConfigFileContent) {
    return (
      baseFields && !!config.config_file_profile && !!config.config_file_content
    );
  }

  return false;
}
