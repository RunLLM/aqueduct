import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import React from 'react';

import ExecutionStatus from '../../utils/shared';

type Props = {
  status: ExecutionStatus;
};

export const Status: React.FC<Props> = ({ status }) => {
  const statusIcons = [];
  const runStatus = status.toLowerCase();

  if (runStatus === ExecutionStatus.Succeeded) {
    statusIcons.push(<Chip label="Succeeded" color="success" size="small" />);
  } else if (runStatus === ExecutionStatus.Failed) {
    statusIcons.push(<Chip label="Failed" color="error" size="small" />);
  } else if (runStatus === ExecutionStatus.Pending) {
    statusIcons.push(<Chip label="In Progress" color="info" size="small" />);
  } else if (runStatus === ExecutionStatus.Canceled) {
    statusIcons.push(<Chip label="Canceled" color="default" size="small" />);
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

export default Status;
