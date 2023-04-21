import Box from '@mui/material/Box';
import React from 'react';

import { LambdaConfig } from '../../../utils/integrations';
import { IntegrationTextInputField } from './IntegrationTextInputField';

// TODO: Add exec_state value, or make this value optional.
const Placeholders: LambdaConfig = {
  role_arn: 'arn:aws:iam::123:role/lambda-function-role-arn',
  exec_state: '',
};

type Props = {
  onUpdateField: (field: keyof LambdaConfig, value: string) => void;
  value?: LambdaConfig;
};

export const LambdaDialog: React.FC<Props> = ({ onUpdateField, value }) => {
  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        name="role_arn"
        spellCheck={false}
        required={true}
        label="Lambda Role ARN"
        description="ARN for Lambda executor role."
        placeholder={Placeholders.role_arn}
        onChange={(event) => onUpdateField('role_arn', event.target.value)}
      />
    </Box>
  );
};

// TODO: Add is Lambda dialog complete function.
