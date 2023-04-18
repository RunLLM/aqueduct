// import { yupResolver } from '@hookform/resolvers/yup';
import Box from '@mui/material/Box';
import TextField from '@mui/material/Textfield';
import Typography from '@mui/material/Typography';
import React from 'react';
import { useFormContext } from 'react-hook-form';

// import * as Yup from 'yup';
import { MariaDbConfig } from '../../../utils/integrations';
import { Button } from '../../primitives/Button.styles';
import { IntegrationTextInputField } from './IntegrationTextInputField';
import { readOnlyFieldWarning, readOnlyFieldDisableReason } from './constants';

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
  editMode,
}) => {
  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        name="host"
        label={'Host*'}
        description={'The hostname or IP address of the MariaDB server.'}
        spellCheck={false}
        required={true}
        placeholder={Placeholders.host}
        onChange={(event) => onUpdateField('host', event.target.value)}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        name="port"
        label={'Port*'}
        description={'The port number of the MariaDB server.'}
        spellCheck={false}
        required={true}
        placeholder={Placeholders.port}
        onChange={(event) => onUpdateField('port', event.target.value)}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        name="database"
        label={'Database*'}
        description={'The name of the specific database to connect to.'}
        spellCheck={false}
        required={true}
        placeholder={Placeholders.database}
        onChange={(event) => onUpdateField('database', event.target.value)}
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
        onChange={(event) => onUpdateField('username', event.target.value)}
      />

      <IntegrationTextInputField
        name="password"
        spellCheck={false}
        required={true}
        label="Password*"
        description="The password corresponding to the above username."
        placeholder={Placeholders.password}
        type="password"
        onChange={(event) => onUpdateField('password', event.target.value)}
      />

    </Box>
  );
};

export const isMariaDBConfigComplete = (config: MariaDbConfig): boolean => {
  return (
    !!config.database &&
    !!config.host &&
    !!config.password &&
    !!config.port &&
    !!config.username &&
    !!config.port
  );
};
