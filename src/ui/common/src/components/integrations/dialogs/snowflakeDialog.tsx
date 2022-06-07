import Box from '@mui/material/Box';
import React, { useEffect, useState } from 'react';

import {
  IntegrationConfig,
  SnowflakeConfig,
} from '../../../utils/integrations';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: SnowflakeConfig = {
  account_identifier: '123456',
  warehouse: 'aqueduct-warehouse',
  database: 'aqueduct-db',
  username: 'aqueduct',
  password: '********',
};

type Props = {
  setDialogConfig: (config: IntegrationConfig) => void;
};

export const SnowflakeDialog: React.FC<Props> = ({ setDialogConfig }) => {
  const [accountIdentifier, setAccountIdentifier] = useState<string>(null);
  const [warehouse, setWarehouse] = useState<string>(null);
  const [database, setDatabase] = useState<string>(null);
  const [username, setUsername] = useState<string>(null);
  const [password, setPassword] = useState<string>(null);

  useEffect(() => {
    const config: SnowflakeConfig = {
      account_identifier: accountIdentifier,
      warehouse: warehouse,
      database: database,
      username: username,
      password: password,
    };
    setDialogConfig(config);
  }, [accountIdentifier, warehouse, database, username, password]);

  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Account Identifier *"
        description="An account identifier for your Snowflake account."
        placeholder={Placeholders.account_identifier}
        onChange={(event) => setAccountIdentifier(event.target.value)}
        value={accountIdentifier}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Warehouse *"
        description="The name of the Snowflake warehouse to connect to."
        placeholder={Placeholders.warehouse}
        onChange={(event) => setWarehouse(event.target.value)}
        value={warehouse}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Database *"
        description="The name of the database to connect to."
        placeholder={Placeholders.database}
        onChange={(event) => setDatabase(event.target.value)}
        value={database}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Username *"
        description="The username of a user with permission to access the database above."
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
