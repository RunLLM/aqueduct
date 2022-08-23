import Box from '@mui/material/Box';
import React from 'react';

import { PostgresConfig } from '../../../utils/integrations';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: PostgresConfig = {
  host: '127.0.0.1',
  port: '5432',
  database: 'aqueduct-db',
  username: 'aqueduct',
  password: '********',
};

type Props = {
  onUpdateField: (field: keyof PostgresConfig, value: string) => void;
  value?: PostgresConfig;
};

export const PostgresDialog: React.FC<Props> = ({ onUpdateField, value }) => {
  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Host *"
        description="The hostname or IP address of the Postgres server."
        placeholder={Placeholders.host}
        onChange={(event) => onUpdateField('host', event.target.value)}
        value={value?.host ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Port *"
        description="The port number of the Postgres server."
        placeholder={Placeholders.port}
        onChange={(event) => onUpdateField('port', event.target.value)}
        value={value?.port ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Database *"
        description="The name of the specific database to connect to."
        placeholder={Placeholders.database}
        onChange={(event) => onUpdateField('database', event.target.value)}
        value={value?.database ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Username *"
        description="The username of a user with access to the above database."
        placeholder={Placeholders.username}
        onChange={(event) => onUpdateField('username', event.target.value)}
        value={value?.username ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={false}
        label="Password"
        description="The password corresponding to the above username."
        placeholder={Placeholders.password}
        type="password"
        onChange={(event) => {
          if (!!event.target.value) {
            onUpdateField('password', event.target.value);
          }
        }}
        value={value?.password ?? null}
      />
    </Box>
  );
};
