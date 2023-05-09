import Box from '@mui/material/Box';
import React from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import {
  IntegrationDialogProps,
  MySqlConfig,
} from '../../../utils/integrations';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: MySqlConfig = {
  host: '127.0.0.1',
  port: '3306',
  database: 'aqueduct-db',
  username: 'aqueduct',
  password: '********',
};

export const MysqlDialog: React.FC<IntegrationDialogProps> = ({
  editMode = false,
}) => {
  const { setValue } = useFormContext();

  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        name="host"
        spellCheck={false}
        required={true}
        label="Host*"
        description="The hostname or IP address of the MySQL server."
        placeholder={Placeholders.host}
        onChange={(event) => setValue('host', event.target.value)}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        name="port"
        spellCheck={false}
        required={true}
        label="Port*"
        description="The port number of the MySQL server."
        placeholder={Placeholders.port}
        onChange={(event) => setValue('port', event.target.value)}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        name="database"
        spellCheck={false}
        required={true}
        label="Database*"
        description="The name of the specific database to connect to."
        placeholder={Placeholders.database}
        onChange={(event) => setValue('database', event.target.value)}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        name="username"
        spellCheck={false}
        required={true}
        label="Username*"
        description="The username of a user with access to the above database."
        placeholder={Placeholders.username}
        onChange={(event) => setValue('username', event.target.value)}
      />

      <IntegrationTextInputField
        name="password"
        spellCheck={false}
        required={true}
        label="Password*"
        description="The password corresponding to the above username."
        placeholder={Placeholders.password}
        type="password"
        onChange={(event) => setValue('password', event.target.value)}
        autoComplete="mysql-password"
      />
    </Box>
  );
};

export function getMySQLValidationSchema() {
  return Yup.object().shape({
    host: Yup.string().required('Please enter a host'),
    // NOTE: backend requires this to be string for now.
    port: Yup.string().required('Please enter a port'),
    database: Yup.string().required('Please enter a database'),
    username: Yup.string().required('Please enter a username'),
    password: Yup.string().required('Please enter a password'),
  });
}
