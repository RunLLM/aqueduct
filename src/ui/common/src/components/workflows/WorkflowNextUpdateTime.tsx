import { faCalendar } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Tooltip, Typography } from '@mui/material';
import React from 'react';

import { WorkflowResponse } from '../../handlers/responses/Workflow';
import { theme } from '../../styles/theme/theme';
import { getNextUpdateTime } from '../../utils/cron';
import { WorkflowUpdateTrigger } from '../../utils/workflows';

type Props = {
  workflow: WorkflowResponse;
};

const WorkflowNextUpdateTime: React.FC<Props> = ({ workflow }) => {
  if (workflow.schedule.trigger !== WorkflowUpdateTrigger.Periodic) {
    return null;
  }

  if (workflow.schedule?.paused) {
    return null;
  }

  if (!workflow.schedule?.cron_schedule) {
    return null;
  }

  const nextUpdateTime = getNextUpdateTime(workflow.schedule?.cron_schedule)
    .toDate()
    .toLocaleString();

  return (
    <Tooltip title="Next Workflow Run" arrow>
      <Box display="flex" alignItems="center" ml={2}>
        <Box mr={1}>
          <FontAwesomeIcon icon={faCalendar} color={theme.palette.gray[800]} />
        </Box>
        <Typography>{nextUpdateTime}</Typography>
      </Box>
    </Tooltip>
  );
};

export default WorkflowNextUpdateTime;
