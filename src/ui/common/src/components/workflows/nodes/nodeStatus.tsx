import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import React from 'react';

import ExecutionStatus, { ExecState, FailureType } from '../../../utils/shared';

type Props = {
  execState: ExecState;
};

export const NodeStatus: React.FC<Props> = ({ execState }) => {
  const statusIcons = [];
  const runStatus = execState.status.toLowerCase();

  if (runStatus === ExecutionStatus.Succeeded) {
    statusIcons.push(<Chip label="Succeeded" color="success" size="small" />);
  } else if (
    runStatus === ExecutionStatus.Failed &&
    execState.failure_type === FailureType.UserNonFatal
  ) {
    statusIcons.push(<Chip label="Warning" color="warning" size="small" />);
  } else if (runStatus === ExecutionStatus.Failed) {
    statusIcons.push(<Chip label="Failed" color="error" size="small" />);
  } else if (runStatus === ExecutionStatus.Pending) {
    statusIcons.push(<Chip label="In Progress" color="default" size="small" />);
  }

  return (
    <Box sx={{ alignItems: 'center' }}>
      {statusIcons.map((icon, idx) => (
        <Box mr={1} key={idx}>
          {icon}
        </Box>
      ))}
    </Box>
  );
};

export default NodeStatus;
