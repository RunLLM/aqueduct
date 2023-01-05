import {
  faCircleCheck,
  faCircleExclamation,
  faCircleQuestion,
  faClockFour,
  faListOl,
  faSpinner,
  faX,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
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
      backgroundColor = theme.palette.Default;
      break;
    case ExecutionStatus.Running:
      backgroundColor = theme.palette.Running;
      break;
    case ExecutionStatus.Succeeded:
      backgroundColor = theme.palette.Success;
      break;
    case ExecutionStatus.Unknown:
    default:
      backgroundColor = theme.palette.Default;
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
      case ExecutionStatus.Unknown:
      case ExecutionStatus.Running:
        textColor = theme.palette.black;
        break;
      case ExecutionStatus.Canceled:
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

type IndicatorProps = {
  /**
   * Execution status to render.
   */
  status: ExecutionStatus;
  /**
   * Size of the Indicator.
   */
  size: string;
  /**
   * False if use default colors. Otherwise, specify the color.
   */
  monochrome: string | false;

};

/**
 * Smaller status indicator component that is just a circle with a background color.
 **/
export const StatusIndicator: React.FC<IndicatorProps> = ({ status, size = '12px', monochrome = false }) => {
  const getIcon = (status: ExecutionStatus) => {
    let indicator = null;
    switch (status) {
      case ExecutionStatus.Running:
        indicator = (
          <Box
            sx={{
              width: '100%',
              height: '100%',
              fontSize: size,
              display: 'flex',
              alignSelf: 'center',
            }}
          >
            <FontAwesomeIcon icon={faSpinner} color={monochrome? monochrome : 'black'} />
          </Box>
        );
        break;

      case ExecutionStatus.Canceled:
        indicator = (
          <Box
            sx={{
              width: '100%',
              height: '100%',
              fontSize: size,
              display: 'flex',
              alignSelf: 'center',
            }}
          >
            <FontAwesomeIcon
              icon={faX}
              color={monochrome? monochrome : getExecutionStatusColor(status)}
            />
          </Box>
        );
        break;
      case ExecutionStatus.Pending:
        indicator = (
          <Box
            sx={{
              width: '100%',
              height: '100%',
              fontSize: size,
              display: 'flex',
              alignSelf: 'center',
            }}
          >
            <FontAwesomeIcon
              icon={faClockFour}
              color={monochrome? monochrome : getExecutionStatusColor(status)}
            />
          </Box>
        );
        break;

      case ExecutionStatus.Registered:
        indicator = (
          <Box
            sx={{
              width: '100%',
              height: '100%',
              fontSize: size,
              display: 'flex',
              alignSelf: 'center',
            }}
          >
            <FontAwesomeIcon
              icon={faListOl}
              color={monochrome? monochrome : getExecutionStatusColor(status)}
            />
          </Box>
        );
        break;

      case ExecutionStatus.Succeeded:
        indicator = (
          <Box
            sx={{
              width: '100%',
              height: '100%',
              fontSize: size,
              display: 'flex',
              alignSelf: 'center',
            }}
          >
            <FontAwesomeIcon
              icon={faCircleCheck}
              color={monochrome? monochrome : getExecutionStatusColor(status)}
            />
          </Box>
        );
        break;

      case ExecutionStatus.Unknown:
        indicator = (
          <Box
            sx={{
              width: '100%',
              height: '100%',
              fontSize: size,
              display: 'flex',
              alignSelf: 'center',
            }}
          >
            <FontAwesomeIcon
              icon={faCircleQuestion}
              color={monochrome? monochrome : getExecutionStatusColor(status)}
            />
          </Box>
        );
        break;

      case ExecutionStatus.Failed:
        indicator = (
          <Box
            sx={{
              width: '100%',
              height: '100%',
              fontSize: size,
              display: 'flex',
              alignSelf: 'center',
            }}
          >
            <FontAwesomeIcon
              icon={faCircleExclamation}
              color={monochrome? monochrome : getExecutionStatusColor(status)}
            />
          </Box>
        );
        break;

      default:
        // No icon, just show a color
        indicator = (
          <Box
            sx={{
              height: '100%',
              width: '100%',
              fontSize: size,
              backgroundColor: getExecutionStatusColor(status),
              borderRadius: 999,
            }}
          />
        );
        break;
    }

    return indicator;
  };

  return (
    <Tooltip title={getExecutionStatusLabel(status)} placement="bottom" arrow>
      <Box
        sx={{
          width: size,
          height: size,
          alignItems: 'center',
          alignSelf: 'center',
        }}
      >
        {getIcon(status)}
      </Box>
    </Tooltip>
  );
};
