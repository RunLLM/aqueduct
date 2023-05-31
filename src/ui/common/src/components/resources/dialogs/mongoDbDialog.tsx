import Box from '@mui/material/Box';
import React from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import { MongoDBConfig, ResourceDialogProps } from '../../../utils/resources';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { ResourceTextInputField } from './ResourceTextInputField';
import { requiredAtCreate } from './schema';

const Placeholders: MongoDBConfig = {
  auth_uri: '********',
  database: 'aqueduct-db',
};

export const MongoDBDialog: React.FC<ResourceDialogProps<MongoDBConfig>> = ({
  resourceToEdit,
}) => {
  const { register, setValue } = useFormContext();

  if (resourceToEdit) {
    Object.entries(resourceToEdit).forEach(([k, v]) => {
      register(k, { value: v });
    });
  }

  const editMode = !!resourceToEdit;

  return (
    <Box sx={{ mt: 2 }}>
      <ResourceTextInputField
        name="auth_uri"
        label={'URI*'}
        description={'The connection URI to your MongoDB server.'}
        spellCheck={false}
        required={true}
        placeholder={Placeholders.auth_uri}
        onChange={(event) => setValue('auth_uri', event.target.value)}
      />

      <ResourceTextInputField
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

export function getMongoDBValidationSchema(editMode: boolean) {
  return Yup.object().shape({
    auth_uri: requiredAtCreate(Yup.string(), editMode, 'Please enter a URI.'),
    database: Yup.string().required('Please enter a database name.'),
  });
}
