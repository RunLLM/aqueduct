import { Box, Typography } from '@mui/material';
import React from 'react';

import { StatusIndicator } from '../../../../components/workflows/workflowStatus';
import ExecutionStatus from '../../../../utils/shared';

interface WorkflowNameItemProps {
  name: string;
  status: ExecutionStatus;
}

export const WorkflowNameItem: React.FC<WorkflowNameItemProps> = ({
  name,
  status,
}) => {
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

export default WorkflowNameItem;
