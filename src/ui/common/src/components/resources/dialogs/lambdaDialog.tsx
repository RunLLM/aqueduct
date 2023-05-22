import Box from '@mui/material/Box';
import React from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import {
  ResourceDialogProps,
  LambdaConfig,
} from '../../../utils/resources';
import { ResourceTextInputField } from './ResourceTextInputField';

const Placeholders: LambdaConfig = {
  role_arn: 'arn:aws:iam::123:role/lambda-function-role-arn',
  exec_state: '',
};

export const LambdaDialog: React.FC<ResourceDialogProps> = ({
  editMode = false,
}) => {
  const { setValue } = useFormContext();
  return (
    <Box sx={{ mt: 2 }}>
      <ResourceTextInputField
        name="role_arn"
        spellCheck={false}
        required={true}
        label="Lambda Role ARN"
        description="ARN for Lambda executor role."
        placeholder={Placeholders.role_arn}
        onChange={(event) => setValue('role_arn', event.target.value)}
      />
    </Box>
  );
};

export function getLambdaValidationSchema() {
  return Yup.object().shape({
    role_arn: Yup.string().required('Please enter a Lambda Role ARN.'),
  });
}
