import Box from '@mui/material/Box';
import Link from '@mui/material/Link';
import Typography from '@mui/material/Typography';
import React from 'react';

import { StatusIndicator } from '../../../components/workflows/workflowStatus';
import ExecutionStatus from '../../../utils/shared';

type DetailsPageHeaderProps = {
  name: string;
  createdAt?: string;
  sourceLocation?: string;
  status: ExecutionStatus;
};

export const DetailsPageHeader: React.FC<DetailsPageHeaderProps> = ({
  name,
  createdAt,
  sourceLocation,
  status,
}) => {
  return (
    <Box width="100%" display="flex" alignItems="center">
      <Box
        sx={{
          width: '24px',
          height: '24px',
          marginRight: '8px',
          display: 'flex',
        }}
      >
        <StatusIndicator status={status} />
      </Box>
      <Typography variant="h4" component="div">
        {name}
      </Typography>

      {createdAt && (
        <Typography marginTop="4px" variant="caption" component="div">
          Created: {createdAt}
        </Typography>
      )}

      {sourceLocation && (
        <Typography variant="caption" component="div">
          Source: <Link>{sourceLocation}</Link>
        </Typography>
      )}
    </Box>
  );
};

export default DetailsPageHeader;
