import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import React from 'react';

import ExecutionStatus, { ExecState, FailureType } from '../../../utils/shared';

type Props = {
  execState: ExecState;
};


// This component seems to be unused. Let's remove it
export const NodeStatus: React.FC<Props> = ({ execState }) => {
  let statusIcon = null;
  const runStatus: string = execState.status.toLowerCase();

  if (runStatus === ExecutionStatus.Succeeded) {
    statusIcon = <Chip label="Succeeded" color="success" size="small" />;
  } else if (
    runStatus === ExecutionStatus.Failed &&
    execState.failure_type === FailureType.UserNonFatal
  ) {
    statusIcon = <Chip label="Warning" color="warning" size="small" />;
  } else if (runStatus === ExecutionStatus.Failed) {
    statusIcon = <Chip label="Failed" color="error" size="small" />;
  } else if (runStatus === ExecutionStatus.Pending) {
    statusIcon = <Chip label="In Progress" color="default" size="small" />;
  } else if (runStatus === ExecutionStatus.Canceled) {
    statusIcon = <Chip label="Canceled" color="default" size="small" />;
  }

  return (
    <Box sx={{ alignItems: 'center' }}>
      <Box mr={1}>{statusIcon}</Box>
    </Box>
  );
};

export default NodeStatus;
