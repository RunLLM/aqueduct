import {
  faCircleCheck,
  faCircleExclamation,
  faCircleQuestion,
  faClockFour,
  faListOl,
  faSpinner,
  faTriangleExclamation,
  faX,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Tooltip } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';

import { theme } from '../../styles/theme/theme';
import ExecutionStatus from '../../utils/shared';

export const getExecutionStatusColor = (status: ExecutionStatus): string => {
  switch (status) {
    case ExecutionStatus.Canceled:
      return theme.palette.Default;
    case ExecutionStatus.Failed:
      return theme.palette.Error;
    case ExecutionStatus.Pending:
      return theme.palette.Info;
    case ExecutionStatus.Registered:
      return theme.palette.Default;
    case ExecutionStatus.Running:
      return theme.palette.Running;
    case ExecutionStatus.Succeeded:
      return theme.palette.Success;
    case ExecutionStatus.Warning:
      return theme.palette.Warning;
    case ExecutionStatus.Unknown:
    default:
      return theme.palette.Default;
  }
};

export const getExecutionStatusLabel = (status: ExecutionStatus): string => {
  switch (status) {
    case ExecutionStatus.Canceled:
      return 'Canceled';
    case ExecutionStatus.Failed:
      return 'Failed';
    case ExecutionStatus.Pending:
      return 'Pending';
    case ExecutionStatus.Registered:
      return 'Registered';
    case ExecutionStatus.Running:
      return 'Running';
    case ExecutionStatus.Succeeded:
      return 'Succeeded';
    case ExecutionStatus.Warning:
      return 'Warning';
    case ExecutionStatus.Unknown:
      return 'Unknown';
    default:
      return 'Unknown';
  }
};

type IndicatorProps = {
  /**
   * Execution status to render.
   */
  status: ExecutionStatus;
  /**
   * Size of the Indicator.
   */
  size?: string;
  /**
   * False if use default colors. Otherwise, specify the color.
   */
  monochrome?: string | false;
  /**
   * Text to show in tooltip. by default, the execution status is shown.
   */
  tooltipText?: string;
  /**
   * If false, no tooltip will be shown. By default, this is true.
   */
  includeTooltip?: boolean;
};

/**
 * Smaller status indicator component that is just a circle with a background color.
 **/
export const StatusIndicator: React.FC<IndicatorProps> = ({
  status,
  size = '20px',
  monochrome = false,
  tooltipText,
  includeTooltip = true,
}) => {
  let icon;
  let spin = false; // Whether or not to spin the icon; this is only used for the running & pending states.
  switch (status) {
    case ExecutionStatus.Running:
      icon = faSpinner;
      spin = true;
      break;

    case ExecutionStatus.Canceled:
      icon = faX;
      break;

    case ExecutionStatus.Pending:
      icon = faSpinner;
      spin = true;
      break;

    case ExecutionStatus.Registered:
      icon = faListOl;
      break;

    case ExecutionStatus.Succeeded:
      icon = faCircleCheck;
      break;

    case ExecutionStatus.Unknown:
      icon = faCircleQuestion;
      break;

    case ExecutionStatus.Failed:
      icon = faCircleExclamation;
      break;

    case ExecutionStatus.Warning:
      icon = faTriangleExclamation;

    default:
      return null; // This can never happen.
  }

  const iconElement = (
    <Box
      sx={{
        width: size,
        height: size,
        alignItems: 'center',
        alignSelf: 'center',
      }}
    >
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
          icon={icon}
          spin={spin}
          color={monochrome ? theme.palette.DarkContrast : getExecutionStatusColor(status)}
        />
      </Box>
    </Box>
  )


  if (!includeTooltip) {
    return iconElement;
  }

  const tooltipTitle = tooltipText || getExecutionStatusLabel(status);

  return (
    <Tooltip
      title={tooltipTitle}
      placement={tooltipText ? 'top' : 'bottom'}
      arrow
    >
      {iconElement}
    </Tooltip>
  );
};
