import Box from '@mui/material/Box';
import React from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import { ResourceDialogProps, SQLiteConfig } from '../../../utils/resources';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { ResourceTextInputField } from './ResourceTextInputField';

const Placeholders: SQLiteConfig = {
  database: '/path/to/sqlite.db',
};

export const SQLiteDialog: React.FC<ResourceDialogProps<SQLiteConfig>> = ({
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
        name="database"
        spellCheck={false}
        required={true}
        label="Path *"
        description="The path to the SQLite file on your Aqueduct server machine."
        placeholder={Placeholders.database}
        onChange={(event) => setValue('database', event.target.value)}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />
    </Box>
  );
};

export function getSQLiteValidationSchema() {
  return Yup.object().shape({
    database: Yup.string().required('Please enter a path'),
  });
}
