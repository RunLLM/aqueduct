import Box from '@mui/material/Box';
import React from 'react';

import { SparkConfig } from '../../../utils/integrations';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: SparkConfig = {
  livy_server_url: 'http://cluster-url.com:8998',
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
        name="livy_server_url"
        label={'Livy Server URL*'}
        description={'URL of Livy Server.'}
        spellCheck={false}
        required={true}
        placeholder={Placeholders.livy_server_url}
        onChange={(event) =>
          onUpdateField('livy_server_url', event.target.value)
        }
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />
    </Box>
  );
};

export function isSparkConfigComplete(config: SparkConfig): boolean {
  return !!config.livy_server_url;
}
