import Box from '@mui/material/Box';
import React, { useState } from 'react';

import { AirflowConfig } from '../../../utils/integrations';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: AirflowConfig = {
  host: 'http://localhost/api/v1',
  username: 'aqueduct',
  password: '********',
  s3_credentials_path: '/home/user/.aws/credentials',
  s3_credentials_profile: 'default',
};

type Props = {
  onUpdateField: (field: keyof AirflowConfig, value: string) => void;
  value?: AirflowConfig;
  editMode: boolean;
};

export const AirflowDialog: React.FC<Props> = ({
  onUpdateField,
  value,
  editMode,
}) => {
  const [host, setHost] = useState<string>(value?.host ?? '');
  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Host *"
        description="The hostname of the Airflow server."
        placeholder={Placeholders.host}
        onChange={(event) => {
          setHost(event.target.value);
          if (event.target.value.startsWith('http://')) {
            // Backend requires the protocol to be stripped
            onUpdateField('host', event.target.value.substring(7));
          } else if (event.target.value.startsWith('https://')) {
            // Backend requires the protocol to be stripped
            onUpdateField('host', event.target.value.substring(8));
          } else {
            onUpdateField('host', event.target.value);
          }
        }}
        value={host}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Username *"
        description="The username of a user with access to the above server."
        placeholder={Placeholders.username}
        onChange={(event) => onUpdateField('username', event.target.value)}
        value={value?.username ?? ''}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Password *"
        description="The password corresponding to the above username."
        placeholder={Placeholders.password}
        type="password"
        onChange={(event) => onUpdateField('password', event.target.value)}
        value={value?.password ?? ''}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="S3 Credentials Path *"
        description="The path on the Airflow server to the AWS credentials that have access to the same S3 bucket configured for Aqueduct storage."
        placeholder={Placeholders.s3_credentials_path}
        onChange={(event) => {
          onUpdateField('s3_credentials_path', event.target.value);
        }}
        value={value?.s3_credentials_path ?? ''}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={false}
        label="S3 Credentials Profile"
        description="The profile to use for the AWS credentials above. The default profile will be used if none is provided."
        placeholder={Placeholders.s3_credentials_profile}
        onChange={(event) => {
          onUpdateField('s3_credentials_profile', event.target.value);
        }}
        value={value?.s3_credentials_profile ?? ''}
      />
    </Box>
  );
};

export function isAirflowConfigComplete(config: AirflowConfig): boolean {
  // required fields:
  // name, host, username, password, s3_credentials_path
  return !!config.host && !!config.username && !!config.password && !!config.s3_credentials_path;
}