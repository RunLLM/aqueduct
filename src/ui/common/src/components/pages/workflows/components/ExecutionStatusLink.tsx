import { Box, Link, Typography } from '@mui/material';
import React from 'react';

import ExecutionStatus from '../../../../utils/shared';
import { StatusIndicator } from '../../../workflows/workflowStatus';

export interface ExecutionStatusLinkProps {
  name: string;
  status: ExecutionStatus;
  url: string;
}

export const ExecutionStatusLink: React.FC<ExecutionStatusLinkProps> = ({
  name,
  status,
  url,
}) => {
  return (
    <Box display="flex" alignItems="left">
      <StatusIndicator status={status} />
      <Link sx={{ cursor: 'pointer' }} href={url}>
        <Typography
          sx={{ marginLeft: '8px', justifyContent: 'right' }}
          variant="body1"
        >
          {name}
        </Typography>
      </Link>
    </Box>
  );
};

export default ExecutionStatusLink;
