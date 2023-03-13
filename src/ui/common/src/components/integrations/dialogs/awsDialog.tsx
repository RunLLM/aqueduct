import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect } from 'react';

import { AWSConfig, AWSCredentialType } from '../../../utils/integrations';
import { Tab, Tabs } from '../../Tabs/Tabs.styles';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: AWSConfig = {
  type: AWSCredentialType.AccessKey,
  region: 'us-east-2',
  access_key_id: '',
  secret_access_key: '',
  config_file_path: '',
  config_file_profile: '',
};

type Props = {
  onUpdateField: (field: keyof AWSConfig, value: string) => void;
  value?: AWSConfig;
  editMode: boolean;
};

export const AWSDialog: React.FC<Props> = ({
  onUpdateField,
  value,
  editMode,
}) => {
  useEffect(() => {
    if (!value?.type) {
      onUpdateField('type', AWSCredentialType.AccessKey);
    }
  }, [onUpdateField, value?.type]);

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
      value={value?.config_file_profile ?? null}
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
        value={value?.access_key_id ?? null}
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
        value={value?.secret_access_key ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="AWS Region*"
        description="The region of your AWS account."
        placeholder={Placeholders.region}
        onChange={(event) => onUpdateField('region', event.target.value)}
        value={value?.region ?? null}
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
        value={value?.config_file_path ?? null}
      />

      {configProfileInput}
    </Box>
  );

  return (
    <Box sx={{ mt: 2 }}>
      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}>
        <Tabs
          value={value?.type}
          onChange={(_, value) => onUpdateField('type', value)}
        >
          <Tab value={AWSCredentialType.AccessKey} label="Enter Access Keys" />
          <Tab
            value={AWSCredentialType.ConfigFilePath}
            label="Specify Path to Credentials"
          />
        </Tabs>
      </Box>
      {value?.type === AWSCredentialType.AccessKey && accessKeyTab}
      {value?.type === AWSCredentialType.ConfigFilePath && configPathTab}
    </Box>
  );
};

export function isAWSConfigComplete(config: AWSConfig): boolean {
  if (config.type === AWSCredentialType.AccessKey) {
    return (
      !!config.access_key_id && !!config.secret_access_key && !!config.region
    );
  }

  if (config.type === AWSCredentialType.ConfigFilePath) {
    return !!config.config_file_profile && !!config.config_file_path;
  }

  return false;
}
