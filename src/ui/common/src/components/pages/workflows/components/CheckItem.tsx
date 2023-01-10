import {
  faCircleCheck,
  faCircleExclamation,
  faCircleXmark,
  faQuestionCircle,
  faTriangleExclamation,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Tooltip, Typography } from '@mui/material';
import React from 'react';
import { useState } from 'react';

import { theme } from '../../../../styles/theme/theme';
import { CheckLevel } from '../../../../utils/operators';
import { ExecutionStatus, showMorePadding } from '../../../../utils/shared';
import { ShowMore } from './MetricItem';

const errorIcon = (
  <Tooltip title="Error" placement="bottom" arrow>
    <Box sx={{ fontSize: '20px', color: theme.palette.red['500'] }}>
      <FontAwesomeIcon icon={faCircleExclamation} />
    </Box>
  </Tooltip>
);

const warningIcon = (
  <Tooltip title="Warning" placement="bottom" arrow>
    <Box sx={{ fontSize: '20px', color: theme.palette.orange['500'] }}>
      <FontAwesomeIcon icon={faTriangleExclamation} />
    </Box>
  </Tooltip>
);

const successIcon = (
  <Tooltip title="Success" placement="bottom" arrow>
    <Box sx={{ fontSize: '20px', color: theme.palette.green['400'] }}>
      <FontAwesomeIcon icon={faCircleCheck} />
    </Box>
  </Tooltip>
);

const unknownIcon = (
  <Tooltip title="Unknown" placement="bottom" arrow>
    <Box sx={{ fontSize: '20px', color: theme.palette.gray['400'] }}>
      <FontAwesomeIcon icon={faQuestionCircle} />
    </Box>
  </Tooltip>
);

const canceledIcon = (
  <Tooltip title="Canceled" placement="bottom" arrow>
    <Box sx={{ fontSize: '20px', color: theme.palette.gray['400'] }}>
      <FontAwesomeIcon icon={faCircleXmark} />
    </Box>
  </Tooltip>
);

export interface CheckPreview {
  checkId: string;
  name: string;
  status: ExecutionStatus;
  level: CheckLevel;
  value?: string;
  // a date.toLocaleString() should go here.
  timestamp: string;
}

interface CheckItemProps {
  checks: CheckPreview[];
}

export const CheckItem: React.FC<CheckItemProps> = ({ checks }) => {
  const [expanded, setExpanded] = useState(false);
  const checksList = [];
  let checksToShow = checks.length;

  if (checks.length > 0) {
    if (!expanded) {
      checksToShow = 1;
    }

    for (let i = 0; i < checksToShow; i++) {
      let statusIcon = successIcon;

      switch (checks[i].status) {
        case ExecutionStatus.Canceled: {
          statusIcon = canceledIcon;
          break;
        }
        case ExecutionStatus.Failed: {
          statusIcon = errorIcon;
          if (checks[i].level === CheckLevel.Warning) {
            statusIcon = warningIcon;
          }
          break;
        }
        case ExecutionStatus.Succeeded: {
          // now we check the value to see if we should show warning or error icon
          if (checks[i].value === 'False') {
            if (checks[i].level === CheckLevel.Error) {
              statusIcon = errorIcon;
            } else {
              statusIcon = warningIcon;
            }
          }
          break;
        }
        case ExecutionStatus.Running:
        case ExecutionStatus.Registered:
        case ExecutionStatus.Pending:
        case ExecutionStatus.Unknown:
        default: {
          statusIcon = unknownIcon;
          break;
        }
      }

      checksList.push(
        <Box
          display="flex"
          key={checks[i].checkId}
          justifyContent="space-between"
          alignItems="center"
        >
          <Typography variant="body1" sx={{ 
            fontWeight: 400,
            color: theme.palette.gray['700'],
            fontSize: '16px',
          }}>
            {checks[i].name}
          </Typography>
          {statusIcon}
        </Box>
      );
    }
  }

  const toggleExpanded = () => {
    setExpanded(!expanded);
  };

  const cellStyling = {
    width: '100%',
  };
  if (checks.length === 1) {
    cellStyling['padding'] = showMorePadding;
  }
  // height 48 because 8px padding top and bottom so 48+2*8=64px
  return (
    <Box
      sx={{
        display: 'flex',
        alignItems: 'center',
        minHeight: '48px',
      }}
    >
      {checks.length > 0 ? (
        <Box sx={cellStyling}>
          {checksList}
          <ShowMore
            totalItems={checks.length}
            numPreviewItems={checksToShow}
            expanded={expanded}
            onClick={toggleExpanded}
          />
        </Box>
      ) : (
        <Typography sx={{ padding: showMorePadding }} variant="body1">
          No checks.
        </Typography>
      )}
    </Box>
  );
};

export default CheckItem;
