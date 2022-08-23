import Box from '@mui/material/Box';
import React from 'react';

import { SnowflakeConfig } from '../../../utils/integrations';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: SnowflakeConfig = {
  account_identifier: '123456',
  warehouse: 'aqueduct-warehouse',
  database: 'aqueduct-db',
  username: 'aqueduct',
  password: '********',
};

type Props = {
  onUpdateField: (field: keyof SnowflakeConfig, value: string) => void;
  value?: SnowflakeConfig;
};

export const SnowflakeDialog: React.FC<Props> = ({ onUpdateField, value }) => {
  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Account Identifier *"
        description="An account identifier for your Snowflake account."
        placeholder={Placeholders.account_identifier}
        onChange={(event) =>
          onUpdateField('account_identifier', event.target.value)
        }
        value={value?.account_identifier ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Warehouse *"
        description="The name of the Snowflake warehouse to connect to."
        placeholder={Placeholders.warehouse}
        onChange={(event) => onUpdateField('warehouse', event.target.value)}
        value={value?.warehouse ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Database *"
        description="The name of the database to connect to."
        placeholder={Placeholders.database}
        onChange={(event) => onUpdateField('database', event.target.value)}
        value={value?.database ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Username *"
        description="The username of a user with permission to access the database above."
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
    </Box>
  );
};
