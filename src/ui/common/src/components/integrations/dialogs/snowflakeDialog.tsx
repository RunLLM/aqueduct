import Box from '@mui/material/Box';
import React, { useEffect, useState } from 'react';

import { SnowflakeConfig } from '../../../utils/integrations';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: SnowflakeConfig = {
  account_identifier: '123456',
  warehouse: 'aqueduct-warehouse',
  database: 'aqueduct-db',
  schema: 'public',
  username: 'aqueduct',
  password: '********',
  role: '',
};

type Props = {
  onUpdateField: (field: keyof SnowflakeConfig, value: string) => void;
  value?: SnowflakeConfig;
  editMode: boolean;
};

export const SnowflakeDialog: React.FC<Props> = ({
  onUpdateField,
  value,
  editMode,
}) => {
  const [schema, setSchema] = useState<string>(
    value?.schema ?? Placeholders.schema
  );

  useEffect(() => {
    if (schema) {
      onUpdateField('schema', schema);
    } else {
      onUpdateField('schema', Placeholders.schema);
    }
  }, [schema]);

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
        warning={editMode ? undefined : readOnlyFieldWarning}
        disabled={editMode}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Warehouse *"
        description="The name of the Snowflake warehouse to connect to."
        placeholder={Placeholders.warehouse}
        onChange={(event) => onUpdateField('warehouse', event.target.value)}
        value={value?.warehouse ?? null}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disabled={editMode}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Database *"
        description="The name of the database to connect to."
        placeholder={Placeholders.database}
        onChange={(event) => onUpdateField('database', event.target.value)}
        value={value?.database ?? null}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disabled={editMode}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={false}
        label="Schema"
        description="The name of the schema to connect to. The public schema will be used if none is provided."
        placeholder={Placeholders.schema}
        onChange={(event) => setSchema(event.target.value)}
        value={schema !== Placeholders.schema ? schema : null}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disabled={editMode}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
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

      <IntegrationTextInputField
        spellCheck={false}
        required={false}
        label="Role"
        description="The role to use when accessing the database above."
        placeholder={Placeholders.role}
        onChange={(event) => onUpdateField('role', event.target.value)}
        value={value?.role ?? null}
      />
    </Box>
  );
};

export function isSnowflakeConfigComplete(config: SnowflakeConfig): boolean {
  const required = [
    'account_identifier',
    'username',
    'password',
    'warehouse',
    'database',
  ];
  return required.every((key) => key in config && config[key].length > 0);
}
