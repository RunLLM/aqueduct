import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import { S3LoadParams } from '../../../utils/operators';

type S3LoadParamDisplayProps = {
  parameters: S3LoadParams;
};

export const S3LoadParamDisplay: React.FC<S3LoadParamDisplayProps> = ({
  parameters,
}) => {
  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'row',
        justifyContent: 'space-evenly',
      }}
    >
      <Box
        sx={{
          textAlign: 'center',
        }}
      >
        <Typography
          variant="body2"
          sx={{
            py: 1,
            color: 'gray.800',
          }}
        >
          Filepath
        </Typography>
        <Typography variant="body1">{parameters.filepath}</Typography>
      </Box>
      <Box
        sx={{
          textAlign: 'center',
        }}
      >
        <Typography
          variant="body2"
          sx={{
            py: 1,
            color: 'gray.800',
          }}
        >
          Format
        </Typography>
        <Typography variant="body1">{parameters.format}</Typography>
      </Box>
    </Box>
  );
};

export default S3LoadParamDisplay;
