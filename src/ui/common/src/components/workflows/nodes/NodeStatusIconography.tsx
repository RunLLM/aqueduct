import { faSkullCrossbones, faXmark } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';

import ExecutionStatus, { ExecState } from '../../../utils/shared';
import { StatusIndicator } from '../workflowStatus';

type Props = {
  execState: ExecState;
  successDisplay: JSX.Element;
};

export const NodeStatusIconography: React.FC<Props> = ({
  execState,
  successDisplay,
}) => {
  if (!execState) {
    return <StatusIndicator status={ExecutionStatus.Pending} size={'50px'} monochrome={'black'}/>;
  } else if (execState.status === ExecutionStatus.Succeeded) {
    return successDisplay;
  } else {
    return <StatusIndicator status={execState.status} size={'50px'} monochrome={'black'} />;
  }
};

export default NodeStatusIconography;
