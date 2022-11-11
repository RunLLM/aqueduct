import { Box, Typography } from '@mui/material';
import React from 'react';

import { StatusIndicator } from '../../../../components/workflows/workflowStatus';
import ExecutionStatus from '../../../../utils/shared';

interface DataNameItemProps {
  name: string;
  status: ExecutionStatus;
}

// TODO: Same as the WorkflowNameItem, may want to consolidate these two components.
export const DataNameItem: React.FC<DataNameItemProps> = ({ name, status }) => {
  return (
    <Box display="flex" alignItems="left">
      <StatusIndicator status={status} />
      <Typography
        sx={{ marginLeft: '8px', justifyContent: 'right' }}
        variant="body1"
      >
        {name}
      </Typography>
    </Box>
  );
};

export default DataNameItem;
