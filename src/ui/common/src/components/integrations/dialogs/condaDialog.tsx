import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';

export const CondaDialog: React.FC = ({}) => {
  return (
    <Box sx={{ mt: 2 }}>
      <Typography variant="body2">
        Before connecting, make sure you have conda installed or have run{' '}
        <code>aqueduct install conda</code>. Once connected, aqueduct server
        will use conda environment to run your future workflows.
      </Typography>
    </Box>
  );
};
