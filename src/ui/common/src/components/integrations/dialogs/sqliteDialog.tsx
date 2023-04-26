import Box from '@mui/material/Box';
import React from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import {
  IntegrationDialogProps,
  SQLiteConfig,
} from '../../../utils/integrations';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: SQLiteConfig = {
  database: '/path/to/sqlite.db',
};

// type Props = {
//   onUpdateField: (field: keyof SQLiteConfig, value: string) => void;
//   value?: SQLiteConfig;
//   editMode: boolean;
// };

export const SQLiteDialog: React.FC<IntegrationDialogProps> = ({
  editMode = false,
}) => {
  const { setValue } = useFormContext();

  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
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

export function isSQLiteConfigComplete(config: SQLiteConfig): boolean {
  return !!config.database;
}

export function getSQLiteValidationSchema() {
  return Yup.object().shape({
    database: Yup.string().required('Please enter a path'),
  });
}
