import Box from '@mui/material/Box';
import React from 'react';

import { MariaDbConfig } from '../../../utils/integrations';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: MariaDbConfig = {
  host: '127.0.0.1',
  port: '3306',
  database: 'aqueduct-db',
  username: 'aqueduct',
  password: '********',
};

type Props = {
  onUpdateField: (field: keyof MariaDbConfig, value: string) => void;
  value?: MariaDbConfig;
  editMode: boolean;
};

export const MariaDbDialog: React.FC<Props> = ({
  onUpdateField,
  value,
  editMode,
}) => {

  console.log('rendering MariaDBDialog ...');

  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        label={'Host*'}
        description={'The hostname or IP address of the MariaDB server.'}
        spellCheck={false}
        required={true}
        placeholder={Placeholders.host}
        onChange={(event) => onUpdateField('host', event.target.value)}
        value={value?.host ?? null}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        label={'Port*'}
        description={'The port number of the MariaDB server.'}
        spellCheck={false}
        required={true}
        placeholder={Placeholders.port}
        onChange={(event) => onUpdateField('port', event.target.value)}
        value={value?.port ?? null}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        label={'Database*'}
        description={'The name of the specific database to connect to.'}
        spellCheck={false}
        required={true}
        placeholder={Placeholders.database}
        onChange={(event) => onUpdateField('database', event.target.value)}
        value={value?.database ?? null}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Username*"
        description="The username of a user with access to the above database."
        placeholder={Placeholders.username}
        onChange={(event) => onUpdateField('username', event.target.value)}
        value={value?.username ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Password*"
        description="The password corresponding to the above username."
        placeholder={Placeholders.password}
        type="password"
        onChange={(event) => onUpdateField('password', event.target.value)}
        value={value?.password ?? null}
      />
    </Box>
  );
};
