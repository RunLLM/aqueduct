import { Tooltip } from '@mui/material';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import React from 'react';

import { theme } from '../../styles/theme/theme';
import ExecutionStatus from '../../utils/shared';

type Props = {
  /**
   * Execution status to render.
   */
  status: ExecutionStatus;
};

export const getExecutionStatusColor = (status: ExecutionStatus): string => {
  let backgroundColor = theme.palette.Primary;
  switch (status) {
    case ExecutionStatus.Canceled:
      backgroundColor = theme.palette.Default;
      break;
    case ExecutionStatus.Failed:
      backgroundColor = theme.palette.Error;
      break;
    case ExecutionStatus.Pending:
      backgroundColor = theme.palette.Info;
      break;
    case ExecutionStatus.Registered:
      backgroundColor = theme.palette.Registered;
      break;
    case ExecutionStatus.Running:
      backgroundColor = theme.palette.Running;
      break;
    case ExecutionStatus.Succeeded:
      backgroundColor = theme.palette.Success;
      break;
    case ExecutionStatus.Unknown:
    default:
      backgroundColor = theme.palette.gray[400];
      break;
  }

  return backgroundColor;
};

export const getExecutionStatusLabel = (status: ExecutionStatus): string => {
  let labelText = 'Succeeded';
  switch (status) {
    case ExecutionStatus.Canceled:
      labelText = 'Canceled';
      break;
    case ExecutionStatus.Failed:
      labelText = 'Failed';
      break;
    case ExecutionStatus.Pending:
      labelText = 'Pending';
      break;
    case ExecutionStatus.Registered:
      labelText = 'Registered';
      break;
    case ExecutionStatus.Running:
      labelText = 'Running';
      break;
    case ExecutionStatus.Succeeded:
      labelText = 'Succeeded';
      break;
    case ExecutionStatus.Unknown:
      labelText = 'Unknown';
      break;
    default:
      labelText = 'Unknown';
      break;
  }

  return labelText;
};

/**
 * Chip component representing an execution status.
 **/
export const StatusChip: React.FC<Props> = ({ status }) => {
  const statusIcons = [];

  const getStatusChipTextColor = (status: ExecutionStatus): string => {
    let textColor = theme.palette.black;
    switch (status) {
      case ExecutionStatus.Canceled:
      case ExecutionStatus.Unknown:
      case ExecutionStatus.Running:
        textColor = theme.palette.black;
        break;
      default:
        textColor = theme.palette.white;
        break;
    }

    return textColor;
  };

  statusIcons.push();

  return (
    <Chip
      label={getExecutionStatusLabel(status)}
      sx={{
        backgroundColor: getExecutionStatusColor(status),
        color: getStatusChipTextColor(status),
      }}
      size="small"
    />
  );
};

export default StatusChip;

/**
 * Smaller status indicator component that is just a circle with a background color.
 **/
export const StatusIndicator: React.FC<Props> = ({ status }) => {
  return (
    <Tooltip title={getExecutionStatusLabel(status)} placement="top" arrow>
      <Box
        sx={{
          width: '12px',
          height: '12px',
          backgroundColor: getExecutionStatusColor(status),
          borderRadius: 999,
          alignSelf: 'center',
        }}
      />
    </Tooltip>
  );
};
