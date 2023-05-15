import { Link, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React, { useEffect } from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import { IntegrationDialogProps } from '../../../utils/integrations';
import { IntegrationTextInputField } from './IntegrationTextInputField';

export const CondaDialog: React.FC<IntegrationDialogProps> = ({
  editMode = false,
}) => {
  const { setValue } = useFormContext();

  useEffect(() => {
    setValue('name', 'Conda');
  });

  return (
    <>
      <IntegrationTextInputField
        name="name"
        spellCheck={false}
        required={true}
        label="Name*"
        description="Provide a unique name to refer to this integration."
        placeholder={'This placeholder should be overwritten.'}
        onChange={(event) => {
          setValue('name', event.target.value);
        }}
        disabled={true}
      />

      <Box sx={{ mt: 2 }}>
        <Typography variant="body2">
          Before connecting, make sure you have{' '}
          <Link
            target="_blank"
            href="https://conda.io/projects/conda/en/latest/user-guide/install/index.html"
          >
            conda
          </Link>{' '}
          and{' '}
          <Link
            target="_blank"
            href="https://conda.io/projects/conda-build/en/latest/install-conda-build.html"
          >
            conda build
          </Link>{' '}
          installed. Once connected, Aqueduct will use conda environments to run
          new workflows.
        </Typography>
      </Box>
    </>
  );
};

export function getCondaValidationSchema() {
  const validationSchema = Yup.object().shape({
    name: Yup.string().required('Please enter a name.'),
  });
  return validationSchema;
}
