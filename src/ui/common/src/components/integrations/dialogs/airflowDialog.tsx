import Box from '@mui/material/Box';
import React, { useEffect, useState } from 'react';

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
  const [address, setAddress] = useState<string>(value?.host ?? null);
  const [s3CredsProfile, setS3CredsProfile] = useState<string>(
    value?.s3_credentials_profile ?? null
  );

  useEffect(() => {
    if (address && address.startsWith('http://')) {
      // Backend requires the protocol to be stripped
      onUpdateField('host', address.substring(7));
    }

    if (address && address.startsWith('https://')) {
      // Backend requires the protocol to be stripped
      onUpdateField('host', address.substring(8));
    }

    if (s3CredsProfile && s3CredsProfile !== 'default') {
      onUpdateField('s3_credentials_profile', s3CredsProfile);
    }
  }, [address, s3CredsProfile]);

  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Host *"
        description="The hostname of the Airflow server."
        placeholder={Placeholders.host}
        onChange={(event) => setAddress(event.target.value)}
        value={address}
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
        value={value?.username ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Password *"
        description="The password corresponding to the above username."
        placeholder={Placeholders.password}
        type="password"
        onChange={(event) => onUpdateField('password', event.target.value)}
        value={value?.password ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="S3 Credentials Path *"
        description="The path on the Airflow server to the AWS credentials that have access to the same S3 bucket configured for Aqueduct storage."
        placeholder={Placeholders.s3_credentials_path}
        onChange={(event) =>
          onUpdateField('s3_credentials_path', event.target.value)
        }
        value={value?.s3_credentials_path ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={false}
        label="S3 Credentials Profile"
        description="The profile to use for the AWS credentials above. The default profile will be used if none is provided."
        placeholder={Placeholders.s3_credentials_profile}
        onChange={(event) => setS3CredsProfile(event.target.value)}
        value={s3CredsProfile}
      />
    </Box>
  );
};
