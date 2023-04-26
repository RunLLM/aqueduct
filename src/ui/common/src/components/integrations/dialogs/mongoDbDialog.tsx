import Box from '@mui/material/Box';
import React from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import {
  IntegrationDialogProps,
  MongoDBConfig,
} from '../../../utils/integrations';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: MongoDBConfig = {
  auth_uri: '********',
  database: 'aqueduct-db',
};

export const MongoDBDialog: React.FC<IntegrationDialogProps> = ({
  editMode = false,
}) => {
  const { setValue } = useFormContext();

  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        name="auth_uri"
        label={'URI*'}
        description={'The connection URI to your MongoDB server.'}
        spellCheck={false}
        required={true}
        placeholder={Placeholders.auth_uri}
        onChange={(event) => setValue('auth_uri', event.target.value)}
      />

      <IntegrationTextInputField
        name="database"
        label={'Database*'}
        description={'The name of the specific database to connect to.'}
        spellCheck={false}
        required={true}
        placeholder={Placeholders.database}
        onChange={(event) => setValue('database', event.target.value)}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />
    </Box>
  );
};

export function isMongoDBConfigComplete(config: MongoDBConfig): boolean {
  return !!config.auth_uri && !!config.database;
}

export function getMongoDBValidationSchema() {
  return Yup.object().shape({
    auth_uri: Yup.string().required('Please enter a URI.'),
    database: Yup.string().required('Please enter a database name.'),
  });
}
