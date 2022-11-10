import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import React from 'react';

import { theme } from '../../styles/theme/theme';
import ExecutionStatus from '../../utils/shared';

type Props = {
  status: ExecutionStatus;
};

export const StatusChip: React.FC<Props> = ({ status }) => {
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
  } else if (runStatus === ExecutionStatus.Registered) {
    statusIcons.push(<Chip label="Pending" color="info" size="small" />);
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

export default StatusChip;

/*
Smaller status indicator component that is just a circle with a color
*/
export const StatusIndicator: React.FC<Props> = ({ status }) => {
  let backgroundColor = theme.palette.Success;
  switch (status) {
    case ExecutionStatus.Canceled:
      backgroundColor = theme.palette.Primary;
      break;
    case ExecutionStatus.Failed:
      backgroundColor = theme.palette.Error;
      break;
    case ExecutionStatus.Pending:
      backgroundColor = theme.palette.Info;
      break;
    case ExecutionStatus.Registered:
      // TODO: Figure out color to use for Registered.
      backgroundColor = theme.palette.Info;
      break;
    case ExecutionStatus.Running:
      // TODO: Figure out color to use for running
      backgroundColor = theme.palette.Info;
      break;
    case ExecutionStatus.Succeeded:
      backgroundColor = theme.palette.Success;
      break;
    case ExecutionStatus.Unknown:
      backgroundColor = theme.palette.gray[400];
      break;
  }

  return (
    <Box
      sx={{
        width: '8px',
        height: '8px',
        backgroundColor,
        borderRadius: 999,
        alignSelf: 'center',
      }}
    />
  );
};
