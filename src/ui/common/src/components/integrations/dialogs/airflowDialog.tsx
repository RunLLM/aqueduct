import Box from '@mui/material/Box';
import React, { useEffect, useState } from 'react';

import { AirflowConfig, IntegrationConfig } from '../../../utils/integrations';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: AirflowConfig = {
  host: 'http://localhost/api/v1',
  username: 'aqueduct',
  password: '********',
  s3_credentials_path: '/home/user/.aws/credentials',
  s3_credentials_profile: 'default',
};

type Props = {
  setDialogConfig: (config: IntegrationConfig) => void;
};

export const AirflowDialog: React.FC<Props> = ({ setDialogConfig }) => {
  const [address, setAddress] = useState<string>(null);
  const [username, setUsername] = useState<string>(null);
  const [password, setPassword] = useState<string>(null);
  const [s3CredsPath, setS3CredsPath] = useState<string>(null);
  const [s3CredsProfile, setS3CredsProfile] = useState<string>(null);

  useEffect(() => {
    const config: AirflowConfig = {
      host: address,
      username: username,
      password: password,
      s3_credentials_path: s3CredsPath,
      s3_credentials_profile: 'default',
    };

    if (address && address.startsWith('http://')) {
      // Backend requires the protocol to be stripped
      config.host = address.substring(7);
    }

    if (address && address.startsWith('https://')) {
      // Backend requires the protocol to be stripped
      config.host = address.substring(8);
    }

    if (s3CredsProfile && s3CredsProfile !== 'default') {
      config.s3_credentials_profile = s3CredsProfile;
    }

    setDialogConfig(config);
  }, [address, username, password, s3CredsPath, s3CredsProfile]);

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
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Username *"
        description="The username of a user with access to the above server."
        placeholder={Placeholders.username}
        onChange={(event) => setUsername(event.target.value)}
        value={username}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Password *"
        description="The password corresponding to the above username."
        placeholder={Placeholders.password}
        type="password"
        onChange={(event) => setPassword(event.target.value)}
        value={password}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="S3 Credentials Path *"
        description="The path on the Airflow server to the AWS credentials that have access to the same S3 bucket configured for Aqueduct storage."
        placeholder={Placeholders.s3_credentials_path}
        onChange={(event) => setS3CredsPath(event.target.value)}
        value={s3CredsPath}
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
