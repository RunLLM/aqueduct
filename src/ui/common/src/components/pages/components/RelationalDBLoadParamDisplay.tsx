import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import { RelationalDBLoadParams } from '../../../utils/operators';
import { InfoTooltip } from './InfoTooltip';

type RelationalDBLoadParamDisplayProps = {
  parameters: RelationalDBLoadParams;
};

export const RelationalDBLoadParamDisplay: React.FC<
  RelationalDBLoadParamDisplayProps
> = ({ parameters }) => {
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
          Table Name
        </Typography>
        <Typography variant="body1">{parameters.table}</Typography>
      </Box>
      <Box
        sx={{
          textAlign: 'center',
        }}
      >
        <Box
          sx={{
            alignItems: 'center',
          }}
        >
          <Typography
            display="inline"
            variant="body2"
            sx={{ color: 'gray.800' }}
          >
            Update Mode
          </Typography>
          <InfoTooltip tooltipText="Action to be taken if the table name already exists" />
        </Box>
        <Typography variant="body1">{parameters.update_mode}</Typography>
      </Box>
    </Box>
  );
};

export default RelationalDBLoadParamDisplay;
