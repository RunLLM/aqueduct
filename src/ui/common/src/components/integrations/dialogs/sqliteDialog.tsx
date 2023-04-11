import Box from '@mui/material/Box';
import React from 'react';

import { SQLiteConfig } from '../../../utils/integrations';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: SQLiteConfig = {
  database: '/path/to/sqlite.db',
};

type Props = {
  onUpdateField: (field: keyof SQLiteConfig, value: string) => void;
  value?: SQLiteConfig;
  editMode: boolean;
};

export const SQLiteDialog: React.FC<Props> = ({
  onUpdateField,
  value,
  editMode,
}) => {
  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        name="path"
        spellCheck={false}
        required={true}
        label="Path *"
        description="The path to the SQLite file on your Aqueduct server machine."
        placeholder={Placeholders.database}
        onChange={(event) => onUpdateField('database', event.target.value)}
        value={value?.database ?? null}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />
    </Box>
  );
};
