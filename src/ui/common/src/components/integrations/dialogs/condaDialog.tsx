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
          conda installed
        </Link>
        . Once connected, aqueduct server will use conda environment to run your
        future workflows.
      </Typography>
    </Box>
  );
};
