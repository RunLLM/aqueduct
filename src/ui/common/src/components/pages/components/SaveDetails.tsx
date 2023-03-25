import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import {
  isGoogleSheetsLoadParams,
  isRelationalDBLoadParams,
  isS3LoadParams,
  LoadParameters,
} from '../../../utils/operators';

type SaveDetailsProps = {
  parameters: LoadParameters;
};

export const SaveDetails: React.FC<SaveDetailsProps> = ({ parameters }) => {
  let paramsDisplay = null;
  if (parameters) {
    if (isRelationalDBLoadParams(parameters)) {
      paramsDisplay = (
        <Box>
          <Box mb={1}>
            <Typography variant="body2" sx={{ color: 'gray.800' }}>
              Table
            </Typography>
            <Typography variant="body1" sx={{ mx: 1 }}>
              {parameters.table}
            </Typography>
          </Box>
          <Box mb={1}>
            <Typography variant="body2" sx={{ color: 'gray.800' }}>
              Update Mode
            </Typography>
            <Typography variant="body1" sx={{ mx: 1 }}>
              {parameters.update_mode}
            </Typography>
          </Box>
        </Box>
      );
    } else if (isGoogleSheetsLoadParams(parameters)) {
      paramsDisplay = (
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
              Save Mode
            </Typography>
            <Typography variant="body1" sx={{ mx: 1 }}>
              {parameters.save_mode}
            </Typography>
          </Box>
        </Box>
      );
    } else if (isS3LoadParams(parameters)) {
      paramsDisplay = (
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
    } else {
      return null;
    }
    return (
      <Box mb={2}>
        <Typography variant="h6" mb="8px" fontWeight="normal">
          Parameters
        </Typography>

        {paramsDisplay}
      </Box>
    );
  }
};

export default SaveDetails;
