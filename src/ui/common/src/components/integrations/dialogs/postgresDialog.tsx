import Box from '@mui/material/Box';
import React from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import {
  IntegrationDialogProps,
  PostgresConfig,
} from '../../../utils/integrations';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: PostgresConfig = {
  host: '127.0.0.1',
  port: '5432',
  database: 'aqueduct-db',
  username: 'aqueduct',
  password: '********',
};

// type Props = {
//   //onUpdateField: (field: keyof PostgresConfig, value: string) => void;
//   //value?: PostgresConfig;
//   editMode: boolean;
// };

export const PostgresDialog: React.FC<IntegrationDialogProps> = ({
  editMode = false,
}) => {
  const { setValue } = useFormContext();

  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        name="host"
        spellCheck={false}
        required={true}
        label="Host *"
        description="The hostname or IP address of the Postgres server."
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
        label="Port *"
        description="The port number of the Postgres server."
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
        label="Database *"
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
        label="Username *"
        description="The username of a user with access to the above database."
        placeholder={Placeholders.username}
        onChange={(event) => setValue('username', event.target.value)}
      />

      <IntegrationTextInputField
        name="password"
        spellCheck={false}
        required={false}
        label="Password"
        description="The password corresponding to the above username."
        placeholder={Placeholders.password}
        type="password"
        onChange={(event) => {
          setValue('password', event.target.value);
        }}
        autoComplete="postgres-password"
      />
    </Box>
  );
};

export function isPostgresConfigComplete(config: PostgresConfig): boolean {
  return (
    !!config.host &&
    !!config.port &&
    !!config.database &&
    !!config.username &&
    !!config.password
  );
}

export function getPostgresValidationSchema() {
  return Yup.object().shape({
    host: Yup.string().required('Please enter a host'),
    port: Yup.string().required('Please enter a port'),
    // to ensure that port is a number:
    // port: Yup.number()
    //   .required('Required')
    //   .typeError('Port must be a number'),
    database: Yup.string().required('Please enter a database'),
    username: Yup.string().required('Please enter a username'),
    password: Yup.string().required('Please enter a password'),
  });
}
