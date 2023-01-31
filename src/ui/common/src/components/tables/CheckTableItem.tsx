import {
  faCircleCheck,
  faCircleExclamation,
  faMinus,
  faTriangleExclamation,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Box from '@mui/material/Box';
import React from 'react';

import { theme } from '../../styles/theme/theme';
import { stringToExecutionStatus } from '../../utils/shared';
import { StatusIndicator } from '../workflows/workflowStatus';

// TODO: Pass in the ExecutionStatus here and render when the check has no value.
interface CheckTableItemProps {
  checkValue: string;
  status?: string;
}

export const CheckTableItem: React.FC<CheckTableItemProps> = ({
  checkValue,
  status,
}) => {
  let iconColor = theme.palette.black;
  let checkIcon = faMinus;

  if (checkValue) {
    switch (checkValue.toLowerCase()) {
      case 'true': {
        checkIcon = faCircleCheck;
        iconColor = theme.palette.Success;
        break;
      }
      case 'false': {
        checkIcon = faCircleExclamation;
        iconColor = theme.palette.Error;
        break;
      }
      case 'warning': {
        checkIcon = faTriangleExclamation;
        iconColor = theme.palette.Warning;
        break;
      }
      case 'none': {
        checkIcon = faMinus;
        iconColor = theme.palette.black;
        break;
      }
      default: {
        // None of the icon cases met, just fall through and render table value.
        return <>{checkValue}</>;
      }
    }
  } else {
    // Check value not found, render the status indicator for this check.
    return (
      <StatusIndicator
        status={stringToExecutionStatus(status)}
        size={'16px'}
        monochrome={false}
      />
    );
  }

  return (
    <Box sx={{ fontSize: '16px', color: iconColor }}>
      <FontAwesomeIcon icon={checkIcon} />
    </Box>
  );
};

export default CheckTableItem;
