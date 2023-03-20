import { Box, Typography } from '@mui/material';
import React from 'react';
import { useState } from 'react';

import { StatusIndicator } from '../../../../components/workflows/workflowStatus';
import { CheckLevel } from '../../../../utils/operators';
import { ExecutionStatus, showMorePadding } from '../../../../utils/shared';
import { ChecksListPreview } from './ChecksListPreview';
import { ShowMore } from './MetricItem';

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

export const getCheckStatusIcon = (
  check: CheckPreview,
  tooltipText?: string
): JSX.Element => {
  const successIcon = (
    <StatusIndicator
      status={ExecutionStatus.Succeeded}
      tooltipText={tooltipText}
    />
  );

  const errorIcon = (
    <StatusIndicator
      status={ExecutionStatus.Failed}
      tooltipText={tooltipText}
    />
  );

  const warningIcon = (
    <StatusIndicator
      status={ExecutionStatus.Warning}
      tooltipText={tooltipText}
    />
  );

  let statusIcon = successIcon;

  switch (check.status) {
    case ExecutionStatus.Succeeded: {
      statusIcon = successIcon;
      break;
    }
    case ExecutionStatus.Failed: {
      if (check.level === CheckLevel.Error) {
        statusIcon = errorIcon;
      } else {
        // CheckLevel.Warning
        statusIcon = warningIcon;
      }

      break;
    }
    default: {
      statusIcon = (
        <StatusIndicator status={check.status} tooltipText={tooltipText} />
      );
      break;
    }
  }
  return statusIcon;
};

export const CheckItem: React.FC<CheckItemProps> = ({ checks }) => {
  const [expanded, setExpanded] = useState(false);
  let checksList = null;
  let checksToShow = checks.length;
  let statusIcon = <StatusIndicator status={ExecutionStatus.Succeeded} />;

  if (checks.length > 0) {
    if (!expanded) {
      checksToShow = 1;
    } else {
      // Initialize empty array to populate checks list.
      checksList = [];
    }

    for (let i = 0; i < checksToShow; i++) {
      statusIcon = getCheckStatusIcon(checks[i]);

      // Show list of checks when expanded.
      // Just show check details if there is one check.
      if (expanded || checks.length === 1) {
        if (checks.length === 1) {
          checksList = [];
        }
        checksList.push(
          <Box
            display="flex"
            key={checks[i].checkId}
            justifyContent="space-between"
            alignItems="center"
          >
            <Typography variant="body1" sx={{ fontWeight: 400 }}>
              {checks[i].name}
            </Typography>
            {statusIcon}
          </Box>
        );
      } else {
        // if contracted, show preview of all checks for the workflow.
        checksList = <ChecksListPreview checks={checks} />;
      }
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
        minWidth: '150px',
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
