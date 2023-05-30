import Box from '@mui/material/Box';
import React from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import { ResourceDialogProps, SnowflakeConfig } from '../../../utils/resources';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { ResourceTextInputField } from './ResourceTextInputField';
import { requiredAtCreate } from './schema';

const Placeholders: SnowflakeConfig = {
  account_identifier: '123456',
  warehouse: 'aqueduct-warehouse',
  database: 'aqueduct-db',
  schema: 'public',
  username: 'aqueduct',
  password: '********',
  role: '',
};

export const SnowflakeDialog: React.FC<
  ResourceDialogProps<SnowflakeConfig>
> = ({ resourceToEdit }) => {
  const { register, setValue } = useFormContext();
  const editMode = !!resourceToEdit;
  if (resourceToEdit) {
    Object.entries(resourceToEdit).forEach(([k, v]) => {
      register(k, { value: v });
    });
  }

  return (
    <Box sx={{ mt: 2 }}>
      <ResourceTextInputField
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

      <ResourceTextInputField
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

      <ResourceTextInputField
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

      <ResourceTextInputField
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

      <ResourceTextInputField
        name="username"
        spellCheck={false}
        required={true}
        label="Username *"
        description="The username of a user with permission to access the database above."
        placeholder={Placeholders.username}
        onChange={(event) => setValue('username', event.target.value)}
      />

      <ResourceTextInputField
        name="password"
        spellCheck={false}
        required={true}
        label="Password *"
        description="The password corresponding to the above username."
        placeholder={Placeholders.password}
        type="password"
        onChange={(event) => setValue('password', event.target.value)}
      />

      <ResourceTextInputField
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

export function getSnowflakeValidationSchema(editMode: boolean) {
  return Yup.object().shape({
    account_identifier: Yup.string().required(
      'Please enter an account identifier'
    ),
    warehouse: Yup.string().required('Please enter a warehouse'),
    database: Yup.string().required('Please enter a database'),
    schema: Yup.string().transform((value) => value || 'public'),
    username: Yup.string().required('Please enter a username'),
    password: requiredAtCreate(
      Yup.string(),
      editMode,
      'Please enter a password'
    ),
    role: Yup.string(),
  });
}
