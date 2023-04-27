import Box from '@mui/material/Box';
import React from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import {
  IntegrationDialogProps,
  SparkConfig,
} from '../../../utils/integrations';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: SparkConfig = {
  livy_server_url: 'http://cluster-url.com:8998',
};

export const SparkDialog: React.FC<IntegrationDialogProps> = ({ editMode }) => {
  const { setValue } = useFormContext();

  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        name="livy_server_url"
        label={'Livy Server URL*'}
        description={'URL of Livy Server.'}
        spellCheck={false}
        required={true}
        placeholder={Placeholders.livy_server_url}
        onChange={(event) => setValue('livy_server_url', event.target.value)}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />
    </Box>
  );
};

export function getSparkValidationSchema() {
  return Yup.object().shape({
    livy_server_url: Yup.string().required('Please enter a Livy Server URL'),
  });
}
