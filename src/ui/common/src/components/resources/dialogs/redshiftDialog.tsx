import Box from '@mui/material/Box';
import React from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import { RedshiftConfig, ResourceDialogProps } from '../../../utils/resources';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { ResourceTextInputField } from './ResourceTextInputField';
import { requiredAtCreate } from './schema';

const Placeholders: RedshiftConfig = {
  host: 'aqueduct.us-east-2.redshift.amazonaws.com',
  port: '5439',
  database: 'aqueduct-db',
  username: 'aqueduct',
  password: '********',
};

export const RedshiftDialog: React.FC<ResourceDialogProps<RedshiftConfig>> = ({
  resourceToEdit,
}) => {
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
        name="host"
        spellCheck={false}
        required={true}
        label="Host *"
        description="The public endpoint of the Redshift cluster."
        placeholder={Placeholders.host}
        onChange={(event) => setValue('host', event.target.value)}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <ResourceTextInputField
        name="port"
        spellCheck={false}
        required={true}
        label="Port *"
        description="The port number of the Redshift cluster."
        placeholder={Placeholders.port}
        onChange={(event) => setValue('port', event.target.value)}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <ResourceTextInputField
        name="database"
        spellCheck={false}
        required={true}
        label="Database *"
        description="The name of the specific database to connect to."
        placeholder={Placeholders.database}
        onChange={(event) => setValue('database', event.target.value)}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <ResourceTextInputField
        name="username"
        spellCheck={false}
        required={true}
        label="Username *"
        description="The username of a user with access to the above database."
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
    </Box>
  );
};

export function getRedshiftValidationSchema(editMode: boolean) {
  return Yup.object().shape({
    host: Yup.string().required('Please enter a host'),
    port: Yup.string().required('Please enter a port'),
    database: Yup.string().required('Please enter a database'),
    username: Yup.string().required('Please enter a username'),
    password: requiredAtCreate(
      Yup.string(),
      editMode,
      'Please enter a password'
    ),
  });
}
