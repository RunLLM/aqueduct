import Box from '@mui/material/Box';
import React from 'react';

import { LambdaConfig } from '../../../utils/integrations';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: LambdaConfig = {
  role_arn: '<my lambda role ARN>',
};

type Props = {
  onUpdateField: (field: keyof LambdaConfig, value: string) => void;
  value?: LambdaConfig;
};

export const LambdaDialog: React.FC<Props> = ({ onUpdateField, value }) => {
  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Lambda Role ARN"
        description="ARN for Lambda executor role."
        placeholder={Placeholders.role_arn}
        onChange={(event) => onUpdateField('role_arn', event.target.value)}
        value={value?.role_arn ?? null}
      />
    </Box>
  );
};
