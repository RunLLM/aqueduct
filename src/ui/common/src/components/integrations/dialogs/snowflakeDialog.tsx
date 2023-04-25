import Box from '@mui/material/Box';
import React from 'react';
import { useFormContext } from 'react-hook-form';

import {
  IntegrationDialogProps,
  SnowflakeConfig,
} from '../../../utils/integrations';
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

// type Props = {
//   onUpdateField: (field: keyof SnowflakeConfig, value: string) => void;
//   value?: SnowflakeConfig;
//   editMode: boolean;
// };

export const SnowflakeDialog: React.FC<IntegrationDialogProps> = ({
  editMode,
}) => {
  const { setValue, getValues } = useFormContext();
  const schema = getValues('schema') ?? Placeholders.schema;
  console.log('snowflakeDialog schema: ', schema);

  // const [schema, setSchema] = useState<string>(
  //   value?.schema ?? Placeholders.schema
  // );

  // useEffect(() => {
  //   if (schema) {
  //     onUpdateField('schema', schema);
  //   } else {
  //     onUpdateField('schema', Placeholders.schema);
  //   }
  // }, [schema]);

  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        name="account_identifier"
        spellCheck={false}
        required={true}
        label="Account Identifier *"
        description="An account identifier for your Snowflake account."
        placeholder={Placeholders.account_identifier}
        onChange={(event) => setValue('account_identifier', event.target.value)}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disabled={editMode}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        name="warehouse"
        spellCheck={false}
        required={true}
        label="Warehouse *"
        description="The name of the Snowflake warehouse to connect to."
        placeholder={Placeholders.warehouse}
        onChange={(event) => setValue('warehouse', event.target.value)}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disabled={editMode}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        name="database"
        spellCheck={false}
        required={true}
        label="Database *"
        description="The name of the database to connect to."
        placeholder={Placeholders.database}
        onChange={(event) => setValue('database', event.target.value)}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disabled={editMode}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        name="schema"
        spellCheck={false}
        required={false}
        label="Schema"
        description="The name of the schema to connect to. The public schema will be used if none is provided."
        placeholder={Placeholders.schema}
        onChange={(event) => setValue('schema', event.target.value)}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disabled={editMode}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        name="username"
        spellCheck={false}
        required={true}
        label="Username *"
        description="The username of a user with permission to access the database above."
        placeholder={Placeholders.username}
        onChange={(event) => setValue('username', event.target.value)}
      />

      <IntegrationTextInputField
        name="password"
        spellCheck={false}
        required={true}
        label="Password *"
        description="The password corresponding to the above username."
        placeholder={Placeholders.password}
        type="password"
        onChange={(event) => setValue('password', event.target.value)}
      />

      <IntegrationTextInputField
        name="role"
        spellCheck={false}
        required={false}
        label="Role"
        description="The role to use when accessing the database above."
        placeholder={Placeholders.role}
        onChange={(event) => setValue('role', event.target.value)}
      />
    </Box>
  );
};

export function isSnowflakeConfigComplete(config: SnowflakeConfig): boolean {
  return (
    !!config.account_identifier &&
    !!config.username &&
    !!config.password &&
    !!config.warehouse &&
    !!config.database
  );
}
