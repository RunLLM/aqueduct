import Box from '@mui/material/Box';
import React, { useEffect, useState } from 'react';

import { IntegrationConfig, RedshiftConfig } from '../../../utils/integrations';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: RedshiftConfig = {
  host: 'aqueduct.us-east-2.redshift.amazonaws.com',
  port: '5439',
  database: 'aqueduct-db',
  username: 'aqueduct',
  password: '********',
};

type Props = {
  setDialogConfig: (config: IntegrationConfig) => void;
};

export const RedshiftDialog: React.FC<Props> = ({ setDialogConfig }) => {
  const [host, setHost] = useState<string>(null);
  const [port, setPort] = useState<string>(null);
  const [database, setDatabase] = useState<string>(null);
  const [username, setUsername] = useState<string>(null);
  const [password, setPassword] = useState<string>(null);

  useEffect(() => {
    const config: RedshiftConfig = {
      host: host,
      port: port,
      database: database,
      username: username,
      password: password,
    };
    setDialogConfig(config);
  }, [host, port, database, username, password]);

  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Host *"
        description="The public endpoint of the Redshift cluster."
        placeholder={Placeholders.host}
        onChange={(event) => setHost(event.target.value)}
        value={host}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Port *"
        description="The port number of the Redshift cluster."
        placeholder={Placeholders.port}
        onChange={(event) => setPort(event.target.value)}
        value={port}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Database *"
        description="The name of the specific database to connect to."
        placeholder={Placeholders.database}
        onChange={(event) => setDatabase(event.target.value)}
        value={database}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Username *"
        description="The username of a user with access to the above database."
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
    </Box>
  );
};
