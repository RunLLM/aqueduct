import Box from '@mui/material/Box';
import React from 'react';
import { useFormContext } from 'react-hook-form';

import {
  AirflowConfig,
  IntegrationDialogProps,
} from '../../../utils/integrations';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: AirflowConfig = {
  host: 'http://localhost/api/v1',
  username: 'aqueduct',
  password: '********',
  s3_credentials_path: '/home/user/.aws/credentials',
  s3_credentials_profile: 'default',
};


export const AirflowDialog: React.FC<IntegrationDialogProps> = ({
  editMode = false,
}) => {
  const { register, setValue } = useFormContext();
  // we need two different values so we can strip the protocol from the host
  register('host');

  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        name="airflow_host" // this value is ignored, host is the value that we're using
        spellCheck={false}
        required={true}
        label="Host *"
        description="The hostname of the Airflow server."
        placeholder={Placeholders.host}
        onChange={(event) => {
          setValue('airflow_host', event.target.value);
          if (event.target.value.startsWith('http://')) {
            // Backend requires the protocol to be stripped
            setValue('host', event.target.value.substring(7));
          } else if (event.target.value.startsWith('https://')) {
            // Backend requires the protocol to be stripped
            setValue('host', event.target.value.substring(8));
          } else {
            setValue('host', event.target.value);
          }
        }}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        name="username"
        spellCheck={false}
        required={true}
        label="Username *"
        description="The username of a user with access to the above server."
        placeholder={Placeholders.username}
        onChange={(event) => setValue('username', event.target.value)}
      />

      <IntegrationTextInputField
        name="password"
        autoComplete="airflow-password"
        spellCheck={false}
        required={true}
        label="Password *"
        description="The password corresponding to the above username."
        placeholder={Placeholders.password}
        type="password"
        onChange={(event) => setValue('password', event.target.value)}
      />

      <IntegrationTextInputField
        name="s3_credentials_path"
        spellCheck={false}
        required={true}
        label="S3 Credentials Path *"
        description="The path on the Airflow server to the AWS credentials that have access to the same S3 bucket configured for Aqueduct storage."
        placeholder={Placeholders.s3_credentials_path}
        onChange={(event) =>
          setValue('s3_credentials_path', event.target.value)
        }
      />

      <IntegrationTextInputField
        name="s3_credentials_profile"
        spellCheck={false}
        required={false}
        label="S3 Credentials Profile"
        description="The profile to use for the AWS credentials above. The default profile will be used if none is provided."
        placeholder={Placeholders.s3_credentials_profile}
        onChange={(event) =>
          setValue('s3_credentials_profile', event.target.value)
        }
      />
    </Box>
  );
};

export function isAirflowConfigComplete(config: AirflowConfig): boolean {
  // required fields:
  // name, host, username, password, s3_credentials_path
  return (
    !!config.host &&
    !!config.username &&
    !!config.password &&
    !!config.s3_credentials_path
  );
}
