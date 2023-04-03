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
    <Box>
      <Box mb={1}>
        <Typography variant="body2" sx={{ color: 'gray.800' }}>
          Filepath
        </Typography>
        <Typography variant="body1" sx={{ mx: 1 }}>
          {parameters.filepath}
        </Typography>
      </Box>
      <Box mb={1}>
        <Typography variant="body2" sx={{ color: 'gray.800' }}>
          Format
        </Typography>
        <Typography variant="body1" sx={{ mx: 1 }}>
          {parameters.format}
        </Typography>
      </Box>
    </Box>
  );
};

export default S3LoadParamDisplay;
