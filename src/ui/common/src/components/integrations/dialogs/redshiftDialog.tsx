import Box from '@mui/material/Box';
import React from 'react';

import { RedshiftConfig } from '../../../utils/integrations';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: RedshiftConfig = {
  host: 'aqueduct.us-east-2.redshift.amazonaws.com',
  port: '5439',
  database: 'aqueduct-db',
  username: 'aqueduct',
  password: '********',
};

type Props = {
  onUpdateField: (field: keyof RedshiftConfig, value: string) => void;
  value?: RedshiftConfig;
  editMode: boolean;
};

export const RedshiftDialog: React.FC<Props> = ({
  onUpdateField,
  value,
  editMode,
}) => {
  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        name="host"
        spellCheck={false}
        required={true}
        label="Host *"
        description="The public endpoint of the Redshift cluster."
        placeholder={Placeholders.host}
        onChange={(event) => onUpdateField('host', event.target.value)}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        name="port"
        spellCheck={false}
        required={true}
        label="Port *"
        description="The port number of the Redshift cluster."
        placeholder={Placeholders.port}
        onChange={(event) => onUpdateField('port', event.target.value)}
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
        onChange={(event) => onUpdateField('database', event.target.value)}
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
        onChange={(event) => onUpdateField('username', event.target.value)}
      />

      <IntegrationTextInputField
        name="password"
        spellCheck={false}
        required={true}
        label="Password *"
        description="The password corresponding to the above username."
        placeholder={Placeholders.password}
        type="password"
        onChange={(event) => onUpdateField('password', event.target.value)}
      />
    </Box>
  );
};

export function isRedshiftConfigComplete(config: RedshiftConfig): boolean {
  return (
    !!config.host &&
    !!config.port &&
    !!config.database &&
    !!config.username &&
    !!config.password
  );
}
