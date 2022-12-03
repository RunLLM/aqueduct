import { Link, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';

export const CondaDialog: React.FC = ({}) => {
  return (
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
  );
};
