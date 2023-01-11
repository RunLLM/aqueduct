import { Box, Typography } from '@mui/material';
import React from 'react';

import ExecutionStatus, { ExecState } from '../../../utils/shared';
import { StatusIndicator } from '../workflowStatus';

type Props = {
  execState: ExecState;
  successDisplay: JSX.Element;
};

export const NodeStatusIconography: React.FC<Props> = ({
  execState,
  successDisplay,
}) => {
  const iconSize = '24px'
  let status = ExecutionStatus.Pending;
  let statusLabel = "fetching";
  if (execState) {
    status = execState.status;
    statusLabel = execState.status;
  }
  if (status === ExecutionStatus.Succeeded) {
    return successDisplay;
  } else {
    return (
      <Box display="flex" alignItems="center">
        <StatusIndicator
          status={status}
          size={iconSize}
          monochrome={'black'}
        />
        <Typography variant="body1" sx={{pl: 1 }}>
          {statusLabel}
        </Typography>
      </Box>     
    );
  }
};

export default NodeStatusIconography;
