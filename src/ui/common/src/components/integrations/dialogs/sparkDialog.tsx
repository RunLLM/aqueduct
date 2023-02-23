import Box from '@mui/material/Box';
import React from 'react';

import { SparkConfig } from '../../../utils/integrations';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: SparkConfig = {
  livy_server_url: '',
};

type Props = {
  onUpdateField: (field: keyof SparkConfig, value: string) => void;
  value?: SparkConfig;
  editMode: boolean;
};

export const SparkDialog: React.FC<Props> = ({
  onUpdateField,
  value,
  editMode,
}) => {
  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        label={'Livy Server URL*'}
        description={'URL of Livy Server.'}
        spellCheck={false}
        required={true}
        placeholder={Placeholders.livy_server_url}
        onChange={(event) =>
          onUpdateField('livy_server_url', event.target.value)
        }
        value={value?.livy_server_url ?? null}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />
    </Box>
  );
};

export function isDatabricksConfigComplete(config: SparkConfig): boolean {
  return !!config.livy_server_url;
}
